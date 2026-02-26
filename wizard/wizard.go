package wizard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/checker/webproject"
	"github.com/mahype/update-watcher/checker/wordpress"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/hostname"
	"github.com/mahype/update-watcher/internal/selfupdate"
	"github.com/mahype/update-watcher/internal/version"
	"github.com/mahype/update-watcher/notifier"
	"github.com/mahype/update-watcher/output"
	"github.com/mahype/update-watcher/runner"
)

// capitalizeFirst returns the string with the first letter uppercased.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// isToolAvailable checks if a command-line tool is on the system PATH.
func isToolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// sudoDescription returns a description for the use_sudo confirm dialog,
// including whether the sudoers file is already configured.
func sudoDescription(command string) string {
	desc := fmt.Sprintf("Required to run '%s' with root privileges.", command)
	if _, err := os.Stat("/etc/sudoers.d/update-watcher"); err == nil {
		desc += " Sudoers file detected."
	} else {
		desc += " Requires /etc/sudoers.d/update-watcher to be configured."
	}
	return desc
}

func sendTestNotification(cfg *config.Config) {
	// Collect enabled notifiers
	var enabled []config.NotifierConfig
	for _, n := range cfg.Notifiers {
		if n.Enabled {
			enabled = append(enabled, n)
		}
	}

	if len(enabled) == 0 {
		fmt.Println("  No enabled notifiers configured.")
		return
	}

	// Build selection options
	var selected string
	var options []huh.Option[string]

	if len(enabled) > 1 {
		options = append(options, huh.NewOption("All notifiers", "__all__"))
	}
	for _, n := range enabled {
		meta, ok := notifier.GetMeta(n.Type)
		label := n.Type
		if ok {
			label = meta.DisplayName
		}
		options = append(options, huh.NewOption(label, n.Type))
	}

	if len(enabled) == 1 {
		selected = enabled[0].Type
	} else {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Send test notification to:").
					Options(options...).
					Value(&selected),
			),
		).Run()
		if err != nil {
			return
		}
	}

	// Build test data
	testResults := []*checker.CheckResult{
		{
			CheckerName: "test",
			Summary:     "2 packages (1 security) — test notification",
			CheckedAt:   time.Now(),
			Updates: []checker.Update{
				{
					Name:           "example-package",
					CurrentVersion: "1.0.0",
					NewVersion:     "1.1.0",
					Type:           checker.UpdateTypeRegular,
				},
				{
					Name:           "libsecurity-example",
					CurrentVersion: "2.0.0",
					NewVersion:     "2.0.1",
					Type:           checker.UpdateTypeSecurity,
					Priority:       checker.PriorityHigh,
				},
			},
		},
	}

	// Determine which notifiers to send to
	var targets []config.NotifierConfig
	if selected == "__all__" {
		targets = enabled
	} else {
		for _, n := range enabled {
			if n.Type == selected {
				targets = append(targets, n)
				break
			}
		}
	}

	// Send test notification(s)
	ctx := context.Background()
	for _, nCfg := range targets {
		meta, ok := notifier.GetMeta(nCfg.Type)
		label := nCfg.Type
		if ok {
			label = meta.DisplayName
		}

		n, err := notifier.Create(nCfg.Type, nCfg)
		if err != nil {
			fmt.Printf("  [!] %s: failed to create notifier: %s\n", label, err)
			continue
		}

		fmt.Printf("  Sending to %s...", label)
		if err := n.Send(ctx, cfg.Hostname, testResults); err != nil {
			fmt.Printf(" failed: %s\n", err)
		} else {
			fmt.Println(" OK")
		}
	}
	fmt.Println()
}

func runTestCheck(cfg *config.Config) {
	// Save config before running so checkers use the current state.
	cfgPath := config.ConfigPath()
	if err := config.Save(cfg, cfgPath); err != nil {
		fmt.Printf("  [!] Failed to save config: %s\n", err)
		return
	}

	fmt.Println("\n  Running test check (no notifications)...")
	fmt.Println()

	noNotify := false
	r := runner.New(cfg, runner.WithNotify(&noNotify))
	result, err := r.Run()
	if err != nil {
		fmt.Printf("  [!] Test check failed: %s\n", err)
		return
	}

	output.PrintResults(result.Results, result.Errors)
	fmt.Println()

	fmt.Print("  Press Enter to continue...")
	fmt.Scanln()
}

const (
	menuWatchers         = "watchers"
	menuNotifications    = "notifications"
	menuSettings         = "settings"
	menuCron             = "cron"
	menuTestRun          = "test"
	menuTestNotification = "test-notification"
	menuSaveExit         = "save"
)

// Run launches the interactive setup wizard.
// If cfg is nil, a fresh default config is used.
func Run(cfg *config.Config) (*config.Config, error) {
	if cfg == nil {
		cfg = config.NewDefault()
	}
	if cfg.Hostname == "" {
		cfg.Hostname = hostname.Get()
	}

	// Check for updates once at startup (non-blocking on error).
	var updateAvailable *selfupdate.Release
	if version.Version != "dev" {
		if release, err := selfupdate.LatestRelease(); err == nil {
			if selfupdate.NeedsUpdate(version.Version, release) {
				updateAvailable = release
			}
		}
	}

	for {
		printStatus(cfg, updateAvailable)

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(buildMainMenuOptions(cfg)...).
					Value(&choice),
			),
		).Run()
		if err != nil {
			// User pressed Ctrl+C — save what we have
			return cfg, nil
		}

		switch choice {
		case menuWatchers:
			if err := manageWatchers(cfg); err != nil {
				return cfg, err
			}
		case menuNotifications:
			if err := manageNotifications(cfg); err != nil {
				return cfg, err
			}
		case menuSettings:
			if err := manageSettings(cfg); err != nil {
				return cfg, err
			}
		case menuCron:
			if err := manageCron(cfg); err != nil {
				return cfg, err
			}
		case menuTestRun:
			runTestCheck(cfg)
		case menuTestNotification:
			sendTestNotification(cfg)
		case menuSaveExit:
			return cfg, nil
		}
	}
}

// ErrTestRun signals that the user wants to run a test check.
var errTestRun = fmt.Errorf("test-run-requested")

// IsTestRunRequested checks if the wizard exit was a test run request.
func IsTestRunRequested(err error) bool {
	return err == errTestRun
}

// formatCronSchedule converts a cron expression to a human-readable string.
// Delegates to cron.FormatSchedule.
func formatCronSchedule(expr string) string {
	return cron.FormatSchedule(expr)
}

// buildMainMenuOptions returns menu options with dynamic labels reflecting current state.
func buildMainMenuOptions(cfg *config.Config) []huh.Option[string] {
	// Watchers label
	watcherLabel := "Manage Watchers"
	if len(cfg.Watchers) == 0 {
		watcherLabel += " (none configured)"
	} else {
		var types []string
		for _, w := range cfg.Watchers {
			t := w.Type
			switch t {
			case "apt", "dnf", "apk":
				t = strings.ToUpper(t)
			case "macos":
				t = "macOS"
			default:
				t = strings.ToUpper(t[:1]) + t[1:]
			}
			types = append(types, t)
		}
		watcherLabel += " (" + strings.Join(types, ", ") + ")"
	}

	// Notifications label
	notifLabel := "Manage Notifications"
	if len(cfg.Notifiers) == 0 {
		notifLabel += " (none configured)"
	} else {
		var names []string
		for _, n := range cfg.Notifiers {
			meta, ok := notifier.GetMeta(n.Type)
			if ok {
				names = append(names, meta.DisplayName)
			} else {
				names = append(names, n.Type)
			}
		}
		notifLabel += " (" + strings.Join(names, ", ") + ")"
	}

	// Settings label
	settingsLabel := "Change Settings"
	if cfg.Hostname != "" {
		settingsLabel += " (hostname: " + cfg.Hostname + ")"
	}

	// Cron label
	cronLabel := "Manage Cron Jobs"
	jobs := cron.InstalledJobs()
	if len(jobs) == 0 {
		cronLabel += " (none)"
	} else {
		var descriptions []string
		for _, j := range jobs {
			descriptions = append(descriptions, fmt.Sprintf("%s: %s",
				cron.JobTypeLabel(j.Type), cron.FormatSchedule(j.Schedule)))
		}
		cronLabel += " (" + strings.Join(descriptions, "; ") + ")"
	}

	opts := []huh.Option[string]{
		huh.NewOption(watcherLabel, menuWatchers),
		huh.NewOption(notifLabel, menuNotifications),
		huh.NewOption(settingsLabel, menuSettings),
		huh.NewOption(cronLabel, menuCron),
		huh.NewOption("Run Test Check", menuTestRun),
	}

	if len(cfg.Notifiers) > 0 {
		opts = append(opts, huh.NewOption("Send Test Notification", menuTestNotification))
	}

	opts = append(opts, huh.NewOption("Save & Exit", menuSaveExit))
	return opts
}

func printStatus(cfg *config.Config, updateAvailable *selfupdate.Release) {
	fmt.Println()
	if version.Version != "dev" {
		fmt.Printf("=== update-watcher setup (%s) ===\n", version.Version)
	} else {
		fmt.Println("=== update-watcher setup ===")
	}
	if updateAvailable != nil {
		fmt.Printf("\033[31m  New version %s available! Run: update-watcher self-update\033[0m\n", updateAvailable.TagName)
	}
	fmt.Println()
	fmt.Printf("  Hostname:      %s\n", cfg.Hostname)

	// Watchers
	if len(cfg.Watchers) == 0 {
		fmt.Println("  Watchers:      (none)")
	} else {
		var names []string
		for _, w := range cfg.Watchers {
			label := w.Type
			if w.Type == "wordpress" {
				sites := w.GetMapSlice("sites")
				label = fmt.Sprintf("wordpress (%d sites)", len(sites))
			}
			if w.Type == "webproject" {
				projects := w.GetMapSlice("projects")
				label = fmt.Sprintf("webproject (%d projects)", len(projects))
			}
			if !w.Enabled {
				label += " [disabled]"
			}
			names = append(names, label)
		}
		fmt.Printf("  Watchers:      %s\n", strings.Join(names, ", "))
	}

	// Notifiers
	if len(cfg.Notifiers) == 0 {
		fmt.Println("  Notifications: (none)")
	} else {
		var names []string
		for _, n := range cfg.Notifiers {
			meta, ok := notifier.GetMeta(n.Type)
			if ok {
				names = append(names, meta.DisplayName)
			} else {
				names = append(names, n.Type)
			}
		}
		fmt.Printf("  Notifications: %s\n", strings.Join(names, ", "))
	}

	// Cron
	cronJobs := cron.InstalledJobs()
	if len(cronJobs) == 0 {
		fmt.Println("  Cron:          not installed")
	} else {
		for i, j := range cronJobs {
			prefix := "  Cron:          "
			if i > 0 {
				prefix = "                 "
			}
			fmt.Printf("%s%s (%s)\n", prefix, cron.JobTypeLabel(j.Type), cron.FormatSchedule(j.Schedule))
		}
	}

	fmt.Printf("  Send policy:   %s\n", cfg.Settings.SendPolicy)
	if cfg.Settings.MinPriority != "" {
		fmt.Printf("  Min priority:  %s\n", cfg.Settings.MinPriority)
	}
	fmt.Println()
}

// --- Watchers sub-menu ---

// watcherSummaryLabel returns a compact label for a watcher in the Level-1 menu.
func watcherSummaryLabel(w config.WatcherConfig) string {
	name := watcherDisplayName(w.Type)
	var details []string

	if !w.Enabled {
		details = append(details, "disabled")
	}

	switch w.Type {
	case "wordpress":
		sites := w.GetMapSlice("sites")
		return fmt.Sprintf("%s (%d sites)", name, len(sites))
	case "webproject":
		projects := w.GetMapSlice("projects")
		return fmt.Sprintf("%s (%d projects)", name, len(projects))
	case "apt":
		if w.GetBool("security_only", false) {
			details = append(details, "security-only")
		}
		if w.GetBool("hide_phased", false) {
			details = append(details, "hide phased")
		}
	case "dnf", "zypper":
		if w.GetBool("security_only", false) {
			details = append(details, "security-only")
		}
	case "macos":
		if w.GetBool("security_only", false) {
			details = append(details, "security-only")
		}
	case "homebrew":
		if w.GetBool("include_casks", true) {
			details = append(details, "with casks")
		}
	case "docker":
		c := w.GetString("containers", "all")
		details = append(details, c)
	case "openclaw":
		if ch := w.GetString("channel", ""); ch != "" {
			details = append(details, ch)
		}
	case "distro":
		if w.GetBool("lts_only", true) {
			details = append(details, "LTS only")
		}
	}

	if len(details) > 0 {
		return fmt.Sprintf("%s (%s)", name, strings.Join(details, ", "))
	}
	return name
}

func manageWatchers(cfg *config.Config) error {
	for {
		// Level 1: List configured watchers
		var options []huh.Option[string]
		for i, w := range cfg.Watchers {
			options = append(options, huh.NewOption(watcherSummaryLabel(w), fmt.Sprintf("select:%d", i)))
		}
		options = append(options,
			huh.NewOption("+ Add new watcher", "add"),
			huh.NewOption("← Back to main menu", "back"),
		)

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Manage Watchers").
					Options(options...).
					Value(&choice),
			),
		).Run()
		if err != nil {
			return nil
		}

		switch {
		case choice == "back":
			return nil
		case choice == "add":
			addWatcherMenu(cfg)
		case strings.HasPrefix(choice, "select:"):
			var idx int
			fmt.Sscanf(choice, "select:%d", &idx)
			if idx >= 0 && idx < len(cfg.Watchers) {
				w := cfg.Watchers[idx]
				switch w.Type {
				case "wordpress":
					watcherSitesMenu(cfg, idx, "sites", "WordPress Sites", "site")
				case "webproject":
					watcherSitesMenu(cfg, idx, "projects", "Web Projects", "project")
				default:
					watcherActionMenu(cfg, idx)
				}
			}
		}
	}
}

// watcherActionMenu shows Edit/Remove options for a single watcher (Level 2).
func watcherActionMenu(cfg *config.Config, idx int) {
	w := cfg.Watchers[idx]
	name := watcherDisplayName(w.Type)

	var options []huh.Option[string]
	if _, ok := editWatcherFuncs[w.Type]; ok {
		options = append(options, huh.NewOption("Edit settings", "edit"))
	}
	options = append(options,
		huh.NewOption("Remove", "remove"),
		huh.NewOption("← Back", "back"),
	)

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(name).
				Options(options...).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	switch choice {
	case "edit":
		if fn, ok := editWatcherFuncs[w.Type]; ok {
			fn(cfg, &cfg.Watchers[idx])
		}
	case "remove":
		removeWatcherByIndex(cfg, idx)
	}
}

// watcherSitesMenu shows sites/projects for a multi-instance watcher (Level 2b).
func watcherSitesMenu(cfg *config.Config, watcherIdx int, itemsKey, title, itemLabel string) {
	for {
		w := &cfg.Watchers[watcherIdx]
		items := w.GetMapSlice(itemsKey)

		var options []huh.Option[string]
		for j, item := range items {
			itemName, _ := item["name"].(string)
			itemPath, _ := item["path"].(string)
			if itemName == "" {
				itemName = "(unnamed)"
			}
			label := fmt.Sprintf("%s (%s)", itemName, itemPath)
			options = append(options, huh.NewOption(label, fmt.Sprintf("select:%d", j)))
		}
		options = append(options,
			huh.NewOption(fmt.Sprintf("+ Add new %s", itemLabel), "add"),
			huh.NewOption("← Back", "back"),
		)

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(title).
					Options(options...).
					Value(&choice),
			),
		).Run()
		if err != nil {
			return
		}

		switch {
		case choice == "back":
			return
		case choice == "add":
			if itemsKey == "sites" {
				addWordPressSite(cfg)
			} else {
				addWebProject(cfg)
			}
			// Re-check: watcher index may have changed if this was the first item
			if watcherIdx >= len(cfg.Watchers) || cfg.Watchers[watcherIdx].Type != w.Type {
				return
			}
		case strings.HasPrefix(choice, "select:"):
			var itemIdx int
			fmt.Sscanf(choice, "select:%d", &itemIdx)
			items = w.GetMapSlice(itemsKey) // refresh
			if itemIdx >= 0 && itemIdx < len(items) {
				watcherSiteActionMenu(cfg, watcherIdx, itemIdx, itemsKey, itemLabel)
				// After action, check if watcher still exists (may have been removed)
				if watcherIdx >= len(cfg.Watchers) {
					return
				}
			}
		}
	}
}

// watcherSiteActionMenu shows Edit/Remove for a single site/project (Level 3).
func watcherSiteActionMenu(cfg *config.Config, watcherIdx, itemIdx int, itemsKey, itemLabel string) {
	w := &cfg.Watchers[watcherIdx]
	items := w.GetMapSlice(itemsKey)
	if itemIdx < 0 || itemIdx >= len(items) {
		return
	}
	itemName, _ := items[itemIdx]["name"].(string)
	if itemName == "" {
		itemName = "(unnamed)"
	}

	title := fmt.Sprintf("%s: %s", watcherDisplayName(w.Type), itemName)

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(
					huh.NewOption("Edit settings", "edit"),
					huh.NewOption("Remove", "remove"),
					huh.NewOption("← Back", "back"),
				).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	switch choice {
	case "edit":
		if itemsKey == "sites" {
			editWordPressSite(cfg, watcherIdx, itemIdx)
		} else {
			editWebProjectEntry(cfg, watcherIdx, itemIdx)
		}
	case "remove":
		removeSiteByIndex(cfg, watcherIdx, itemIdx, itemsKey, itemLabel)
	}
}

// addWatcherMenu shows available watcher types to add (Add submenu).
func addWatcherMenu(cfg *config.Config) {
	var options []huh.Option[string]

	if runtime.GOOS == "linux" && isToolAvailable("apt") {
		options = append(options, huh.NewOption("APT", "add-apt"))
	}
	if runtime.GOOS == "linux" && isToolAvailable("dnf") {
		options = append(options, huh.NewOption("DNF", "add-dnf"))
	}
	if runtime.GOOS == "linux" && isToolAvailable("pacman") {
		options = append(options, huh.NewOption("Pacman", "add-pacman"))
	}
	if runtime.GOOS == "linux" && isToolAvailable("zypper") {
		options = append(options, huh.NewOption("Zypper", "add-zypper"))
	}
	if runtime.GOOS == "linux" && isToolAvailable("apk") {
		options = append(options, huh.NewOption("APK", "add-apk"))
	}
	if runtime.GOOS == "darwin" && isToolAvailable("softwareupdate") {
		options = append(options, huh.NewOption("macOS", "add-macos"))
	}
	if isToolAvailable("brew") {
		options = append(options, huh.NewOption("Homebrew", "add-homebrew"))
	}
	if isToolAvailable("snap") {
		options = append(options, huh.NewOption("Snap", "add-snap"))
	}
	if isToolAvailable("npm") {
		options = append(options, huh.NewOption("npm", "add-npm"))
	}
	if isToolAvailable("flatpak") {
		options = append(options, huh.NewOption("Flatpak", "add-flatpak"))
	}
	if isToolAvailable("docker") {
		options = append(options, huh.NewOption("Docker", "add-docker"))
	}
	options = append(options, huh.NewOption("OpenClaw", "add-openclaw"))
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/etc/os-release"); err == nil {
			options = append(options, huh.NewOption("Distro Release", "add-distro"))
		}
	}
	options = append(options, huh.NewOption("Web Project", "add-webproject"))
	options = append(options, huh.NewOption("WordPress", "add-wordpress"))
	options = append(options, huh.NewOption("← Back", "back"))

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Add Watcher").
				Options(options...).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	switch choice {
	case "add-apt":
		addAptWatcher(cfg)
	case "add-dnf":
		addDnfWatcher(cfg)
	case "add-pacman":
		addPacmanWatcher(cfg)
	case "add-zypper":
		addZypperWatcher(cfg)
	case "add-apk":
		addApkWatcher(cfg)
	case "add-macos":
		addMacOSWatcher(cfg)
	case "add-homebrew":
		addHomebrewWatcher(cfg)
	case "add-snap":
		addSnapWatcher(cfg)
	case "add-npm":
		addNpmWatcher(cfg)
	case "add-flatpak":
		addFlatpakWatcher(cfg)
	case "add-docker":
		addDockerWatcher(cfg)
	case "add-openclaw":
		addOpenClawWatcher(cfg)
	case "add-distro":
		addDistroWatcher(cfg)
	case "add-wordpress":
		addWordPressSite(cfg)
	case "add-webproject":
		addWebProject(cfg)
	}
}

// removeWatcherByIndex removes a watcher at the given index after confirmation.
func removeWatcherByIndex(cfg *config.Config, idx int) {
	name := watcherDisplayName(cfg.Watchers[idx].Type)
	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Remove %s watcher?", name)).
				Value(&confirm),
		),
	).Run()
	if err != nil || !confirm {
		return
	}
	cfg.Watchers = append(cfg.Watchers[:idx], cfg.Watchers[idx+1:]...)
	fmt.Printf("  %s watcher removed.\n", name)
}

// removeSiteByIndex removes a site/project at the given index from a multi-instance watcher.
func removeSiteByIndex(cfg *config.Config, watcherIdx, itemIdx int, itemsKey, itemLabel string) {
	w := &cfg.Watchers[watcherIdx]
	items := w.GetMapSlice(itemsKey)
	if itemIdx < 0 || itemIdx >= len(items) {
		return
	}

	itemName, _ := items[itemIdx]["name"].(string)
	if itemName == "" {
		itemName = "(unnamed)"
	}

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Remove %s %q?", itemLabel, itemName)).
				Value(&confirm),
		),
	).Run()
	if err != nil || !confirm {
		return
	}

	if len(items) == 1 {
		// Last item — remove the entire watcher entry.
		cfg.Watchers = append(cfg.Watchers[:watcherIdx], cfg.Watchers[watcherIdx+1:]...)
		fmt.Printf("  %s %q removed (watcher removed, no %ss remaining).\n", capitalizeFirst(itemLabel), itemName, itemLabel)
	} else {
		remaining := make([]interface{}, 0, len(items)-1)
		for k, s := range items {
			if k != itemIdx {
				remaining = append(remaining, s)
			}
		}
		w.Options[itemsKey] = remaining
		fmt.Printf("  %s %q removed.\n", capitalizeFirst(itemLabel), itemName)
	}
}

func addAptWatcher(cfg *config.Config) {
	securityOnly := false
	useSudo := true
	hidePhased := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "apt" {
			securityOnly = w.GetBool("security_only", false)
			useSudo = w.GetBool("use_sudo", true)
			hidePhased = w.GetBool("hide_phased", false)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report security updates?").
				Value(&securityOnly),
			huh.NewConfirm().
				Title("Use sudo for apt operations?").
				Description(sudoDescription("apt-get update")).
				Value(&useSudo),
			huh.NewConfirm().
				Title("Hide phased updates?").
				Description("Phased updates are gradually rolled out by Ubuntu and cannot be installed immediately. Hide them to reduce noise.").
				Value(&hidePhased),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "apt",
		Enabled: true,
		Options: map[string]interface{}{
			"security_only": securityOnly,
			"use_sudo":      useSudo,
			"hide_phased":   hidePhased,
		},
	})
	fmt.Println("  APT watcher configured.")
}

func addDnfWatcher(cfg *config.Config) {
	securityOnly := false
	useSudo := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "dnf" {
			securityOnly = w.GetBool("security_only", false)
			useSudo = w.GetBool("use_sudo", true)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report security updates?").
				Value(&securityOnly),
			huh.NewConfirm().
				Title("Use sudo for dnf operations?").
				Description(sudoDescription("dnf check-update")).
				Value(&useSudo),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "dnf",
		Enabled: true,
		Options: map[string]interface{}{
			"security_only": securityOnly,
			"use_sudo":      useSudo,
		},
	})
	fmt.Println("  DNF watcher configured.")
}

func addPacmanWatcher(cfg *config.Config) {
	useSudo := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "pacman" {
			useSudo = w.GetBool("use_sudo", true)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use sudo for pacman sync operations?").
				Description(sudoDescription("pacman -Sy")).
				Value(&useSudo),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "pacman",
		Enabled: true,
		Options: map[string]interface{}{
			"use_sudo": useSudo,
		},
	})
	fmt.Println("  Pacman watcher configured.")

	if !isToolAvailable("arch-audit") {
		fmt.Println()
		fmt.Println("  Tip: Install arch-audit for security update detection:")
		fmt.Println("     pacman -S arch-audit")
	}
}

func addZypperWatcher(cfg *config.Config) {
	securityOnly := false
	useSudo := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "zypper" {
			securityOnly = w.GetBool("security_only", false)
			useSudo = w.GetBool("use_sudo", true)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report security updates?").
				Value(&securityOnly),
			huh.NewConfirm().
				Title("Use sudo for zypper operations?").
				Description(sudoDescription("zypper refresh")).
				Value(&useSudo),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "zypper",
		Enabled: true,
		Options: map[string]interface{}{
			"security_only": securityOnly,
			"use_sudo":      useSudo,
		},
	})
	fmt.Println("  Zypper watcher configured.")
}

func addApkWatcher(cfg *config.Config) {
	useSudo := false

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "apk" {
			useSudo = w.GetBool("use_sudo", false)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use sudo for apk operations?").
				Description(sudoDescription("apk update")).
				Value(&useSudo),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "apk",
		Enabled: true,
		Options: map[string]interface{}{
			"use_sudo": useSudo,
		},
	})
	fmt.Println("  APK watcher configured.")
}

func addMacOSWatcher(cfg *config.Config) {
	securityOnly := false

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "macos" {
			securityOnly = w.GetBool("security_only", false)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report security updates?").
				Value(&securityOnly),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "macos",
		Enabled: true,
		Options: map[string]interface{}{
			"security_only": securityOnly,
		},
	})
	fmt.Println("  macOS watcher configured.")
}

func addHomebrewWatcher(cfg *config.Config) {
	includeCasks := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "homebrew" {
			includeCasks = w.GetBool("include_casks", true)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include cask updates?").
				Description("Casks are macOS GUI applications (e.g. Firefox, VS Code)").
				Value(&includeCasks),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "homebrew",
		Enabled: true,
		Options: map[string]interface{}{
			"include_casks": includeCasks,
		},
	})
	fmt.Println("  Homebrew watcher configured.")
}

func addSnapWatcher(cfg *config.Config) {
	cfg.AddWatcher(config.WatcherConfig{
		Type:    "snap",
		Enabled: true,
		Options: map[string]interface{}{},
	})
	fmt.Println("  Snap watcher configured.")
}

func addNpmWatcher(cfg *config.Config) {
	cfg.AddWatcher(config.WatcherConfig{
		Type:    "npm",
		Enabled: true,
		Options: map[string]interface{}{},
	})
	fmt.Println("  npm global watcher configured.")
}

func addFlatpakWatcher(cfg *config.Config) {
	cfg.AddWatcher(config.WatcherConfig{
		Type:    "flatpak",
		Enabled: true,
		Options: map[string]interface{}{},
	})
	fmt.Println("  Flatpak watcher configured.")
}

func addOpenClawWatcher(cfg *config.Config) {
	detected := isToolAvailable("openclaw")
	binaryPath := ""
	channel := ""

	// Pre-fill from existing watcher config.
	for _, w := range cfg.Watchers {
		if w.Type == "openclaw" {
			channel = w.GetString("channel", "")
			binaryPath = w.GetString("binary_path", "")
			break
		}
	}

	useManualPath := !detected
	if detected {
		autoUse := binaryPath == "" // default: use auto if no manual path stored
		huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("openclaw found in PATH. Use detected binary?").
					Description("Select 'No' to specify a custom binary path instead.").
					Value(&autoUse),
			),
		).Run()
		if !autoUse {
			useManualPath = true
		}
	}

	if useManualPath {
		if binaryPath == "" {
			binaryPath = "openclaw"
		}
		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Path to openclaw binary").
					Description("Run 'which openclaw' to locate it, or try:\nfind /usr /opt /home -name openclaw 2>/dev/null | head -5").
					Value(&binaryPath),
				huh.NewSelect[string]().
					Title("Update channel").
					Description("Which OpenClaw update channel to monitor").
					Options(
						huh.NewOption("Stable (default)", ""),
						huh.NewOption("Beta", "beta"),
						huh.NewOption("Dev", "dev"),
					).
					Value(&channel),
			),
		).Run()
	} else {
		huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Update channel").
					Description("Which OpenClaw update channel to monitor").
					Options(
						huh.NewOption("Stable (default)", ""),
						huh.NewOption("Beta", "beta"),
						huh.NewOption("Dev", "dev"),
					).
					Value(&channel),
			),
		).Run()
	}

	opts := map[string]interface{}{}
	if channel != "" {
		opts["channel"] = channel
	}
	if useManualPath && binaryPath != "" && binaryPath != "openclaw" {
		opts["binary_path"] = binaryPath
	}
	cfg.AddWatcher(config.WatcherConfig{
		Type:    "openclaw",
		Enabled: true,
		Options: opts,
	})
	fmt.Println("  OpenClaw watcher configured.")
}

func addDistroWatcher(cfg *config.Config) {
	ltsOnly := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "distro" {
			ltsOnly = w.GetBool("lts_only", true)
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report LTS upgrades?").
				Description("Ubuntu only: Skip short-lived releases (e.g. 23.10) and only report upgrades to the next Long Term Support version (e.g. 22.04 → 24.04). Has no effect on Debian or Fedora.").
				Value(&ltsOnly),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "distro",
		Enabled: true,
		Options: map[string]interface{}{
			"lts_only": ltsOnly,
		},
	})
	fmt.Println("  Distro release watcher configured.")
}

func addDockerWatcher(cfg *config.Config) {
	containers := "all"
	excludeStr := ""

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "docker" {
			containers = w.GetString("containers", "all")
			exclude := w.GetStringSlice("exclude", nil)
			excludeStr = strings.Join(exclude, ", ")
			break
		}
	}

	huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Containers to check").
				Description("\"all\" or comma-separated names").
				Value(&containers),
			huh.NewInput().
				Title("Containers to exclude").
				Description("Comma-separated names (leave empty for none)").
				Value(&excludeStr),
		),
	).Run()

	options := map[string]interface{}{
		"containers": containers,
	}
	if excludeStr != "" {
		parts := strings.Split(excludeStr, ",")
		exclude := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				exclude = append(exclude, t)
			}
		}
		options["exclude"] = exclude
	}

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "docker",
		Enabled: true,
		Options: options,
	})
	fmt.Println("  Docker watcher configured.")
}

func addWordPressSite(cfg *config.Config) {
	var siteName, sitePath string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Site name").
				Description("A human-readable name for this site").
				Value(&siteName),
			huh.NewInput().
				Title("WordPress path").
				Description("Full path to WordPress project root, e.g. /var/www/html/blog or ~/Dev/Sites/mysite").
				Value(&sitePath),
		),
	).Run()
	if err != nil || sitePath == "" {
		fmt.Println("  Skipped (path is required).")
		return
	}

	if siteName == "" {
		siteName = sitePath
	}

	// Auto-detect environment
	detectedEnv := wordpress.DetectEnvironment(sitePath)
	envStr := string(detectedEnv)

	fmt.Printf("\n  Detected environment: %s (%s)\n", detectedEnv.Label(), wordpress.EnvironmentDescription(detectedEnv))

	// Let user confirm or change the environment
	envOptions := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("%s (detected)", detectedEnv.Label()), string(detectedEnv)),
	}
	for _, e := range wordpress.AllEnvironments {
		if e != detectedEnv {
			envOptions = append(envOptions, huh.NewOption(e.Label(), string(e)))
		}
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Environment").
				Description("How WP-CLI should be invoked for this site").
				Options(envOptions...).
				Value(&envStr),
		),
	).Run()
	if err != nil {
		fmt.Println("  Cancelled.")
		return
	}

	selectedEnv := wordpress.Environment(envStr)

	site := map[string]interface{}{
		"name":        siteName,
		"path":        sitePath,
		"environment": envStr,
	}

	// Only ask for run_as if the environment needs it (native)
	if selectedEnv.NeedsRunAs() {
		runAs := "www-data"
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Run WP-CLI as user").
					Description("OS user for sudo -u (only needed for native installs)").
					Value(&runAs),
			),
		).Run()
		if err == nil && runAs != "" {
			site["run_as"] = runAs
		}
	}

	// Find existing wordpress watcher or create new one
	var found bool
	for i, w := range cfg.Watchers {
		if w.Type == "wordpress" {
			sites := w.GetMapSlice("sites")
			sites = append(sites, site)
			sitesIface := make([]interface{}, len(sites))
			for j, s := range sites {
				sitesIface[j] = s
			}
			if cfg.Watchers[i].Options == nil {
				cfg.Watchers[i].Options = make(map[string]interface{})
			}
			cfg.Watchers[i].Options["sites"] = sitesIface
			found = true
			break
		}
	}

	if !found {
		cfg.AddWatcher(config.WatcherConfig{
			Type:    "wordpress",
			Enabled: true,
			Options: map[string]interface{}{
				"sites":         []interface{}{site},
				"check_core":    true,
				"check_plugins": true,
				"check_themes":  true,
			},
		})
	}

	fmt.Printf("  WordPress site %q added (env: %s).\n", siteName, selectedEnv.Label())
}

func addWebProject(cfg *config.Config) {
	var projectName, projectPath string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("A human-readable name for this project").
				Value(&projectName),
			huh.NewInput().
				Title("Project path").
				Description("Full path to project root, e.g. /var/www/myapp or ~/Dev/Projects/myproject").
				Value(&projectPath),
		),
	).Run()
	if err != nil || projectPath == "" {
		fmt.Println("  Skipped (path is required).")
		return
	}

	if projectName == "" {
		projectName = projectPath
	}

	// Auto-detect environment
	detectedEnv := webproject.DetectEnvironment(projectPath)
	envStr := string(detectedEnv)

	fmt.Printf("\n  Detected environment: %s (%s)\n", detectedEnv.Label(), webproject.EnvironmentDescription(detectedEnv))

	// Let user confirm or change the environment
	envOptions := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("%s (detected)", detectedEnv.Label()), string(detectedEnv)),
	}
	for _, e := range webproject.AllEnvironments {
		if e != detectedEnv {
			envOptions = append(envOptions, huh.NewOption(e.Label(), string(e)))
		}
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Environment").
				Description("How package manager commands should be invoked").
				Options(envOptions...).
				Value(&envStr),
		),
	).Run()
	if err != nil {
		fmt.Println("  Cancelled.")
		return
	}

	selectedEnv := webproject.Environment(envStr)

	// Auto-detect package managers
	detected := webproject.DetectManagers(projectPath)
	var managerNames []string
	for _, m := range detected {
		managerNames = append(managerNames, m.Name())
	}
	if len(managerNames) > 0 {
		fmt.Printf("  Detected package managers: %s\n", strings.Join(managerNames, ", "))
	} else {
		fmt.Println("  No package managers detected (will be auto-detected at runtime).")
	}

	// Security audit option
	checkAudit := true
	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Run security audits?").
				Description("Check for known vulnerabilities (npm audit, composer audit, etc.)").
				Value(&checkAudit),
		),
	).Run()

	project := map[string]interface{}{
		"name":        projectName,
		"path":        projectPath,
		"environment": envStr,
		"check_audit": checkAudit,
	}

	if len(managerNames) > 0 {
		mgrs := make([]interface{}, len(managerNames))
		for i, m := range managerNames {
			mgrs[i] = m
		}
		project["managers"] = mgrs
	}

	// Only ask for run_as if the environment needs it (native)
	if selectedEnv.NeedsRunAs() {
		runAs := ""
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Run as user").
					Description("OS user for sudo -u (leave empty to skip)").
					Value(&runAs),
			),
		).Run()
		if err == nil && runAs != "" {
			project["run_as"] = runAs
		}
	}

	// Find existing webproject watcher or create new one
	var found bool
	for i, w := range cfg.Watchers {
		if w.Type == "webproject" {
			projects := w.GetMapSlice("projects")
			projects = append(projects, project)
			projectsIface := make([]interface{}, len(projects))
			for j, p := range projects {
				projectsIface[j] = p
			}
			if cfg.Watchers[i].Options == nil {
				cfg.Watchers[i].Options = make(map[string]interface{})
			}
			cfg.Watchers[i].Options["projects"] = projectsIface
			found = true
			break
		}
	}

	if !found {
		cfg.AddWatcher(config.WatcherConfig{
			Type:    "webproject",
			Enabled: true,
			Options: map[string]interface{}{
				"check_audit": checkAudit,
				"projects":    []interface{}{project},
			},
		})
	}

	fmt.Printf("  Web project %q added (env: %s", projectName, selectedEnv.Label())
	if len(managerNames) > 0 {
		fmt.Printf(", managers: %s", strings.Join(managerNames, ", "))
	}
	fmt.Println(").")
}


// --- Notifications sub-menu ---

// notifierSummaryLabel returns a compact label for a notifier in the Level-1 menu.
func notifierSummaryLabel(n config.NotifierConfig) string {
	meta, ok := notifier.GetMeta(n.Type)
	name := n.Type
	if ok {
		name = meta.DisplayName
	}
	var details []string
	if !n.Enabled {
		details = append(details, "disabled")
	}
	if n.SendPolicy != "" {
		details = append(details, n.SendPolicy)
	}
	if n.MinPriority != "" {
		details = append(details, n.MinPriority+"+")
	}
	if len(details) > 0 {
		return fmt.Sprintf("%s (%s)", name, strings.Join(details, ", "))
	}
	return name
}

func manageNotifications(cfg *config.Config) error {
	for {
		// Level 1: List configured notifiers
		var options []huh.Option[string]
		for i, n := range cfg.Notifiers {
			options = append(options, huh.NewOption(notifierSummaryLabel(n), fmt.Sprintf("select:%d", i)))
		}
		options = append(options,
			huh.NewOption("+ Add new notifier", "add"),
			huh.NewOption("← Back to main menu", "back"),
		)

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Manage Notifications").
					Options(options...).
					Value(&choice),
			),
		).Run()
		if err != nil {
			return nil
		}

		switch {
		case choice == "back":
			return nil
		case choice == "add":
			addNotifierMenu(cfg)
		case strings.HasPrefix(choice, "select:"):
			var idx int
			fmt.Sscanf(choice, "select:%d", &idx)
			if idx >= 0 && idx < len(cfg.Notifiers) {
				notifierActionMenu(cfg, idx)
			}
		}
	}
}

// notifierActionMenu shows Edit/Remove options for a single notifier (Level 2).
func notifierActionMenu(cfg *config.Config, idx int) {
	meta, ok := notifier.GetMeta(cfg.Notifiers[idx].Type)
	name := cfg.Notifiers[idx].Type
	if ok {
		name = meta.DisplayName
	}

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(name).
				Options(
					huh.NewOption("Edit settings", "edit"),
					huh.NewOption("Remove", "remove"),
					huh.NewOption("← Back", "back"),
				).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	switch choice {
	case "edit":
		n := &cfg.Notifiers[idx]
		if fn, ok := editFuncs[n.Type]; ok {
			fn(cfg, n)
		}
	case "remove":
		removeNotifierByIndex(cfg, idx)
	}
}

// addNotifierMenu shows available notifier types to add.
func addNotifierMenu(cfg *config.Config) {
	var options []huh.Option[string]
	for _, meta := range notifier.AllMeta() {
		options = append(options, huh.NewOption(meta.DisplayName, "add:"+meta.Type))
	}
	options = append(options, huh.NewOption("← Back", "back"))

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Add Notifier").
				Options(options...).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	if strings.HasPrefix(choice, "add:") {
		notifierType := strings.TrimPrefix(choice, "add:")
		if fn, ok := addFuncs[notifierType]; ok {
			fn(cfg)
		}
	}
}

// removeNotifierByIndex removes a notifier at the given index after confirmation.
func removeNotifierByIndex(cfg *config.Config, idx int) {
	meta, ok := notifier.GetMeta(cfg.Notifiers[idx].Type)
	name := cfg.Notifiers[idx].Type
	if ok {
		name = meta.DisplayName
	}

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Remove %s notifier?", name)).
				Value(&confirm),
		),
	).Run()
	if err != nil || !confirm {
		return
	}

	cfg.Notifiers = append(cfg.Notifiers[:idx], cfg.Notifiers[idx+1:]...)
	fmt.Printf("  %s notifier removed.\n", name)
}

// --- Settings sub-menu ---

func manageSettings(cfg *config.Config) error {
	hostnameVal := cfg.Hostname
	sendPolicy := cfg.Settings.SendPolicy
	minPriority := cfg.Settings.MinPriority

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Hostname").
				Description(fmt.Sprintf("Auto-detected: %s", hostname.Get())).
				Value(&hostnameVal),
			huh.NewSelect[string]().
				Title("Send policy").
				Options(
					huh.NewOption("Only when updates are found", "only-on-updates"),
					huh.NewOption("Always (even when no updates)", "always"),
				).
				Value(&sendPolicy),
			huh.NewSelect[string]().
				Title("Minimum priority (global)").
				Description("Only send updates at or above this priority. Per-notifier settings override this.").
				Options(
					huh.NewOption("No filter (all updates)", ""),
					huh.NewOption("Low and above", "low"),
					huh.NewOption("Normal and above", "normal"),
					huh.NewOption("High and above", "high"),
					huh.NewOption("Critical only", "critical"),
				).
				Value(&minPriority),
		),
	).Run()
	if err != nil {
		return nil
	}

	cfg.Hostname = hostnameVal
	cfg.Settings.SendPolicy = sendPolicy
	cfg.Settings.MinPriority = minPriority
	fmt.Println("  Settings updated.")
	return nil
}

// --- Cron sub-menu ---

func manageCron(cfg *config.Config) error {
	for {
		jobs := cron.InstalledJobs()

		// Display current jobs
		fmt.Println()
		fmt.Println("  Scheduled cron jobs:")
		if len(jobs) == 0 {
			fmt.Println("    (none)")
		}
		for _, j := range jobs {
			fmt.Printf("    %s — %s\n", cron.JobTypeLabel(j.Type), cron.FormatSchedule(j.Schedule))
		}
		fmt.Println()

		// Build menu options
		checkInstalled, _ := cron.IsJobInstalled(cron.JobCheck)
		selfUpdateInstalled, _ := cron.IsJobInstalled(cron.JobSelfUpdate)

		var options []huh.Option[string]

		if !checkInstalled {
			options = append(options, huh.NewOption("Add update check schedule", "add-check"))
		} else {
			options = append(options, huh.NewOption("Change update check schedule", "edit-check"))
			options = append(options, huh.NewOption("Remove update check schedule", "remove-check"))
		}

		if !selfUpdateInstalled {
			options = append(options, huh.NewOption("Add self-update schedule", "add-self-update"))
		} else {
			options = append(options, huh.NewOption("Change self-update schedule", "edit-self-update"))
			options = append(options, huh.NewOption("Remove self-update schedule", "remove-self-update"))
		}

		if len(jobs) > 0 {
			options = append(options, huh.NewOption("Remove all cron jobs", "remove-all"))
		}
		options = append(options, huh.NewOption("Back to main menu", "back"))

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Manage Cron Jobs").
					Options(options...).
					Value(&choice),
			),
		).Run()
		if err != nil {
			return nil
		}

		switch choice {
		case "add-check", "edit-check":
			installCronJobInteractive(cfg, cron.JobCheck)
		case "add-self-update", "edit-self-update":
			installCronJobInteractive(cfg, cron.JobSelfUpdate)
		case "remove-check":
			if err := cron.UninstallJob(cron.JobCheck); err != nil {
				fmt.Printf("  Error: %s\n", err)
			} else {
				cfg.RemoveCronJob(config.CronJobCheck)
				fmt.Println("  Update check cron job removed.")
			}
		case "remove-self-update":
			if err := cron.UninstallJob(cron.JobSelfUpdate); err != nil {
				fmt.Printf("  Error: %s\n", err)
			} else {
				cfg.RemoveCronJob(config.CronJobSelfUpdate)
				fmt.Println("  Self-update cron job removed.")
			}
		case "remove-all":
			if err := cron.UninstallAll(); err != nil {
				fmt.Printf("  Error: %s\n", err)
			} else {
				cfg.Settings.CronJobs = nil
				fmt.Println("  All cron jobs removed.")
			}
		case "back":
			return nil
		}
	}
}

func installCronJobInteractive(cfg *config.Config, jobType cron.JobType) {
	// Schedule type selection
	var scheduleType string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Schedule type for %s", cron.JobTypeLabel(jobType))).
				Options(
					huh.NewOption("Daily at a specific time (e.g. 07:00)", "daily"),
					huh.NewOption("Interval (e.g. every 6 hours)", "interval"),
					huh.NewOption("Custom cron expression", "custom"),
				).
				Value(&scheduleType),
		),
	).Run()
	if err != nil {
		return
	}

	var cronExpr string

	switch scheduleType {
	case "daily":
		cronTime := "07:00"
		// Pre-fill from existing job
		if existing := cfg.FindCronJob(config.CronJobType(jobType)); existing != nil {
			parts := strings.Fields(existing.Schedule)
			if len(parts) >= 2 {
				if m, err1 := strconv.Atoi(parts[0]); err1 == nil {
					if h, err2 := strconv.Atoi(parts[1]); err2 == nil {
						cronTime = fmt.Sprintf("%02d:%02d", h, m)
					}
				}
			}
		}

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Check time (HH:MM)").
					Value(&cronTime),
			),
		).Run()
		if err != nil {
			return
		}

		hour, minute, parseErr := cron.ParseTime(cronTime)
		if parseErr != nil {
			fmt.Printf("  Error: %s\n", parseErr)
			return
		}
		cronExpr = fmt.Sprintf("%d %d * * *", minute, hour)

	case "interval":
		intervalValue := "6"
		intervalUnit := "hours"

		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Interval value").
					Description("How often to run (number)").
					Value(&intervalValue),
				huh.NewSelect[string]().
					Title("Interval unit").
					Options(
						huh.NewOption("Hours", "hours"),
						huh.NewOption("Minutes", "minutes"),
					).
					Value(&intervalUnit),
			),
		).Run()
		if err != nil {
			return
		}

		val, parseErr := strconv.Atoi(intervalValue)
		if parseErr != nil {
			fmt.Printf("  Error: invalid number %q\n", intervalValue)
			return
		}

		cronExpr, err = cron.IntervalToExpr(val, intervalUnit)
		if err != nil {
			fmt.Printf("  Error: %s\n", err)
			return
		}

	case "custom":
		cronExpr = "0 7 * * *"
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Cron expression").
					Description("Five-field cron format: minute hour day month weekday").
					Value(&cronExpr),
			),
		).Run()
		if err != nil {
			return
		}
	}

	// Install in crontab
	if err := cron.InstallJobWithExpr(jobType, cronExpr); err != nil {
		fmt.Printf("  Error: %s\n", err)
		fmt.Printf("  You can install manually: update-watcher install-cron --type=%s --cron-expr=%q\n", jobType, cronExpr)
		return
	}

	// Save to config
	cfg.AddCronJob(config.CronJob{
		Type:     config.CronJobType(jobType),
		Schedule: cronExpr,
	})

	fmt.Printf("  %s cron job installed: %s\n", cron.JobTypeLabel(jobType), cron.FormatSchedule(cronExpr))
}
