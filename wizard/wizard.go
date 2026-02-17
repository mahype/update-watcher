package wizard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mahype/update-watcher/checker/wordpress"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/cron"
	"github.com/mahype/update-watcher/internal/hostname"
	"github.com/mahype/update-watcher/notifier"
)

// isToolAvailable checks if a command-line tool is on the system PATH.
func isToolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

const (
	menuWatchers      = "watchers"
	menuNotifications = "notifications"
	menuSettings      = "settings"
	menuCron          = "cron"
	menuTestRun       = "test"
	menuSaveExit      = "save"
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

	for {
		printStatus(cfg)

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
			return cfg, errTestRun
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
func formatCronSchedule(expr string) string {
	parts := strings.Fields(expr)
	if len(parts) < 5 {
		return expr
	}
	if parts[2] == "*" && parts[3] == "*" && parts[4] == "*" {
		minute, mErr := strconv.Atoi(parts[0])
		hour, hErr := strconv.Atoi(parts[1])
		if mErr == nil && hErr == nil {
			return fmt.Sprintf("daily at %02d:%02d", hour, minute)
		}
	}
	return expr
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
	notifLabel := "Configure Notifications"
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
	cronLabel := "Manage Cron Job"
	installed, schedule := cron.IsInstalled()
	if installed {
		cronLabel += " (" + formatCronSchedule(schedule) + ")"
	} else {
		cronLabel += " (not installed)"
	}

	return []huh.Option[string]{
		huh.NewOption(watcherLabel, menuWatchers),
		huh.NewOption(notifLabel, menuNotifications),
		huh.NewOption(settingsLabel, menuSettings),
		huh.NewOption(cronLabel, menuCron),
		huh.NewOption("Run Test Check", menuTestRun),
		huh.NewOption("Save & Exit", menuSaveExit),
	}
}

func printStatus(cfg *config.Config) {
	fmt.Println()
	fmt.Println("=== update-watcher setup ===")
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
	installed, schedule := cron.IsInstalled()
	if installed {
		fmt.Printf("  Cron:          %s\n", schedule)
	} else {
		fmt.Println("  Cron:          not installed")
	}

	fmt.Printf("  Send policy:   %s\n", cfg.Settings.SendPolicy)
	fmt.Println()
}

// --- Watchers sub-menu ---

func manageWatchers(cfg *config.Config) error {
	for {
		fmt.Println()
		fmt.Println("  Configured watchers:")
		if len(cfg.Watchers) == 0 {
			fmt.Println("    (none)")
		}
		for _, w := range cfg.Watchers {
			status := "enabled"
			if !w.Enabled {
				status = "disabled"
			}
			switch w.Type {
			case "apt":
				secOnly := w.GetBool("security_only", false)
				fmt.Printf("    [✓] APT (%s, security_only: %v)\n", status, secOnly)
			case "dnf":
				secOnly := w.GetBool("security_only", false)
				fmt.Printf("    [✓] DNF (%s, security_only: %v)\n", status, secOnly)
			case "pacman":
				fmt.Printf("    [✓] Pacman (%s)\n", status)
			case "zypper":
				secOnly := w.GetBool("security_only", false)
				fmt.Printf("    [✓] Zypper (%s, security_only: %v)\n", status, secOnly)
			case "apk":
				fmt.Printf("    [✓] APK (%s)\n", status)
			case "macos":
				secOnly := w.GetBool("security_only", false)
				fmt.Printf("    [✓] macOS (%s, security_only: %v)\n", status, secOnly)
			case "docker":
				containers := w.GetString("containers", "all")
				fmt.Printf("    [✓] Docker (%s, containers: %s)\n", status, containers)
			case "wordpress":
				sites := w.GetMapSlice("sites")
				for _, s := range sites {
					name, _ := s["name"].(string)
					path, _ := s["path"].(string)
					env, _ := s["environment"].(string)
					if env == "" || env == "auto" {
						env = "native"
					}
					fmt.Printf("    [✓] WordPress: %q (%s, env: %s)\n", name, path, env)
				}
			}
		}
		fmt.Println()

		var options []huh.Option[string]
		// Only show options for tools that are available on this system.
		if runtime.GOOS == "linux" && isToolAvailable("apt") {
			options = append(options, huh.NewOption("Add APT watcher", "add-apt"))
		}
		if runtime.GOOS == "linux" && isToolAvailable("dnf") {
			options = append(options, huh.NewOption("Add DNF watcher", "add-dnf"))
		}
		if runtime.GOOS == "linux" && isToolAvailable("pacman") {
			options = append(options, huh.NewOption("Add Pacman watcher", "add-pacman"))
		}
		if runtime.GOOS == "linux" && isToolAvailable("zypper") {
			options = append(options, huh.NewOption("Add Zypper watcher", "add-zypper"))
		}
		if runtime.GOOS == "linux" && isToolAvailable("apk") {
			options = append(options, huh.NewOption("Add APK watcher", "add-apk"))
		}
		if runtime.GOOS == "darwin" && isToolAvailable("softwareupdate") {
			options = append(options, huh.NewOption("Add macOS watcher", "add-macos"))
		}
		if isToolAvailable("docker") {
			options = append(options, huh.NewOption("Add Docker watcher", "add-docker"))
		}
		// WordPress is always available (environment detection handles tool requirements).
		options = append(options, huh.NewOption("Add WordPress site", "add-wordpress"))
		if len(cfg.Watchers) > 0 {
			options = append(options, huh.NewOption("Remove a watcher", "remove"))
		}
		options = append(options, huh.NewOption("Back to main menu", "back"))

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
		case "add-docker":
			addDockerWatcher(cfg)
		case "add-wordpress":
			addWordPressSite(cfg)
		case "remove":
			removeWatcher(cfg)
		case "back":
			return nil
		}
	}
}

func addAptWatcher(cfg *config.Config) {
	securityOnly := false
	useSudo := true

	// Pre-fill from existing
	for _, w := range cfg.Watchers {
		if w.Type == "apt" {
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
				Title("Use sudo for apt operations?").
				Value(&useSudo),
		),
	).Run()

	cfg.AddWatcher(config.WatcherConfig{
		Type:    "apt",
		Enabled: true,
		Options: map[string]interface{}{
			"security_only": securityOnly,
			"use_sudo":      useSudo,
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

func removeWatcher(cfg *config.Config) {
	if len(cfg.Watchers) == 0 {
		return
	}

	var options []huh.Option[string]
	for i, w := range cfg.Watchers {
		if w.Type == "wordpress" {
			// Show one option per WordPress site for individual removal.
			sites := w.GetMapSlice("sites")
			for j, s := range sites {
				name, _ := s["name"].(string)
				path, _ := s["path"].(string)
				env, _ := s["environment"].(string)
				if env == "" || env == "auto" {
					env = "native"
				}
				label := fmt.Sprintf("WordPress: %q (%s, env: %s)", name, path, env)
				value := fmt.Sprintf("wp:%d:%d", i, j)
				options = append(options, huh.NewOption(label, value))
			}
		} else {
			// Non-WordPress watchers: one option to remove the whole watcher.
			label := w.Type
			switch w.Type {
			case "apt":
				label = "APT"
			case "macos":
				label = "macOS"
			default:
				label = strings.ToUpper(label[:1]) + label[1:]
			}
			options = append(options, huh.NewOption(label, fmt.Sprintf("idx:%d", i)))
		}
	}
	options = append(options, huh.NewOption("Cancel", "cancel"))

	var choice string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which watcher to remove?").
				Options(options...).
				Value(&choice),
		),
	).Run()
	if err != nil {
		return
	}

	if choice == "cancel" {
		return
	}

	if strings.HasPrefix(choice, "wp:") {
		// WordPress site removal: parse "wp:<watcherIdx>:<siteIdx>"
		var watcherIdx, siteIdx int
		fmt.Sscanf(choice, "wp:%d:%d", &watcherIdx, &siteIdx)
		if watcherIdx < 0 || watcherIdx >= len(cfg.Watchers) {
			return
		}
		w := &cfg.Watchers[watcherIdx]
		sites := w.GetMapSlice("sites")
		if siteIdx < 0 || siteIdx >= len(sites) {
			return
		}

		removedName, _ := sites[siteIdx]["name"].(string)

		if len(sites) == 1 {
			// Last site — remove the entire WordPress watcher entry.
			cfg.Watchers = append(cfg.Watchers[:watcherIdx], cfg.Watchers[watcherIdx+1:]...)
			fmt.Printf("  WordPress site %q removed (watcher removed, no sites remaining).\n", removedName)
		} else {
			// Remove just this site from the sites slice.
			remaining := make([]interface{}, 0, len(sites)-1)
			for k, s := range sites {
				if k != siteIdx {
					remaining = append(remaining, s)
				}
			}
			w.Options["sites"] = remaining
			fmt.Printf("  WordPress site %q removed.\n", removedName)
		}
	} else if strings.HasPrefix(choice, "idx:") {
		// Non-WordPress watcher removal: parse "idx:<watcherIdx>"
		var idx int
		fmt.Sscanf(choice, "idx:%d", &idx)
		if idx >= 0 && idx < len(cfg.Watchers) {
			removed := cfg.Watchers[idx].Type
			cfg.Watchers = append(cfg.Watchers[:idx], cfg.Watchers[idx+1:]...)
			fmt.Printf("  %s watcher removed.\n", removed)
		}
	}
}

// --- Notifications sub-menu ---

func manageNotifications(cfg *config.Config) error {
	for {
		// Show configured notifiers
		fmt.Println()
		fmt.Println("  Configured notifiers:")
		if len(cfg.Notifiers) == 0 {
			fmt.Println("    (none)")
		}
		for _, n := range cfg.Notifiers {
			status := "enabled"
			if !n.Enabled {
				status = "disabled"
			}
			meta, ok := notifier.GetMeta(n.Type)
			displayName := n.Type
			if ok {
				displayName = meta.DisplayName
			}
			fmt.Printf("    [✓] %s (%s)\n", displayName, status)
		}
		fmt.Println()

		// Build menu options
		var options []huh.Option[string]

		// Add options for each available notifier type
		for _, meta := range notifier.AllMeta() {
			options = append(options, huh.NewOption(
				fmt.Sprintf("Add %s", meta.DisplayName),
				"add:"+meta.Type,
			))
		}

		// Edit/remove options for configured notifiers
		for i, n := range cfg.Notifiers {
			meta, ok := notifier.GetMeta(n.Type)
			displayName := n.Type
			if ok {
				displayName = meta.DisplayName
			}
			options = append(options,
				huh.NewOption(fmt.Sprintf("Edit %s", displayName), fmt.Sprintf("edit:%d", i)),
				huh.NewOption(fmt.Sprintf("Remove %s", displayName), fmt.Sprintf("remove:%d", i)),
			)
		}

		options = append(options, huh.NewOption("Back to main menu", "back"))

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

		if choice == "back" {
			return nil
		}

		if strings.HasPrefix(choice, "add:") {
			notifierType := strings.TrimPrefix(choice, "add:")
			if fn, ok := addFuncs[notifierType]; ok {
				if err := fn(cfg); err != nil {
					return err
				}
			}
		} else if strings.HasPrefix(choice, "edit:") {
			var idx int
			fmt.Sscanf(choice, "edit:%d", &idx)
			if idx >= 0 && idx < len(cfg.Notifiers) {
				n := &cfg.Notifiers[idx]
				if fn, ok := editFuncs[n.Type]; ok {
					if err := fn(cfg, n); err != nil {
						return err
					}
				}
			}
		} else if strings.HasPrefix(choice, "remove:") {
			var idx int
			fmt.Sscanf(choice, "remove:%d", &idx)
			if idx >= 0 && idx < len(cfg.Notifiers) {
				meta, ok := notifier.GetMeta(cfg.Notifiers[idx].Type)
				displayName := cfg.Notifiers[idx].Type
				if ok {
					displayName = meta.DisplayName
				}
				cfg.Notifiers = append(cfg.Notifiers[:idx], cfg.Notifiers[idx+1:]...)
				fmt.Printf("  %s notifier removed.\n", displayName)
			}
		}
	}
}

// --- Settings sub-menu ---

func manageSettings(cfg *config.Config) error {
	hostnameVal := cfg.Hostname
	sendPolicy := cfg.Settings.SendPolicy

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
		),
	).Run()
	if err != nil {
		return nil
	}

	cfg.Hostname = hostnameVal
	cfg.Settings.SendPolicy = sendPolicy
	fmt.Println("  Settings updated.")
	return nil
}

// --- Cron sub-menu ---

func manageCron(cfg *config.Config) error {
	installed, schedule := cron.IsInstalled()

	if installed {
		fmt.Printf("\n  Cron job installed: %s\n\n", schedule)

		var choice string
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Cron Job").
					Options(
						huh.NewOption("Change schedule", "change"),
						huh.NewOption("Remove cron job", "remove"),
						huh.NewOption("Back to main menu", "back"),
					).
					Value(&choice),
			),
		).Run()
		if err != nil {
			return nil
		}

		switch choice {
		case "change":
			return installCronInteractive(cfg)
		case "remove":
			if err := cron.Uninstall(); err != nil {
				fmt.Printf("  Error: %s\n", err)
			} else {
				fmt.Println("  Cron job removed.")
			}
		}
	} else {
		var install bool
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Install cron job for daily checks?").
					Value(&install),
			),
		).Run()
		if err != nil {
			return nil
		}
		if install {
			return installCronInteractive(cfg)
		}
	}
	return nil
}

func installCronInteractive(cfg *config.Config) error {
	cronTime := "07:00"
	if cfg.Settings.Schedule != "" {
		// Try to extract time from existing cron expression (e.g. "0 7 * * *" -> "07:00")
		parts := strings.Fields(cfg.Settings.Schedule)
		if len(parts) >= 2 {
			cronTime = fmt.Sprintf("%02s:%02s", parts[1], parts[0])
		}
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Check time (HH:MM)").
				Value(&cronTime),
		),
	).Run()
	if err != nil {
		return nil
	}

	if err := cron.Install(cronTime); err != nil {
		fmt.Printf("  Error: %s\n", err)
		fmt.Printf("  You can install manually: update-watcher install-cron --time=%s\n", cronTime)
	} else {
		fmt.Printf("  Cron job installed for daily checks at %s\n", cronTime)
	}
	return nil
}
