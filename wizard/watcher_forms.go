package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mahype/update-watcher/checker/webproject"
	"github.com/mahype/update-watcher/checker/wordpress"
	"github.com/mahype/update-watcher/config"
)

// editWatcherFuncs maps watcher types to their edit-configuration functions.
var editWatcherFuncs = map[string]func(cfg *config.Config, existing *config.WatcherConfig) error{
	"apt":      editAptWatcher,
	"dnf":      editDnfWatcher,
	"pacman":   editPacmanWatcher,
	"zypper":   editZypperWatcher,
	"apk":      editApkWatcher,
	"macos":    editMacOSWatcher,
	"homebrew": editHomebrewWatcher,
	"docker":   editDockerWatcher,
	"openclaw": editOpenClawWatcher,
	"distro":   editDistroWatcher,
}

// watcherDisplayName returns a human-readable name for the given watcher type.
func watcherDisplayName(typ string) string {
	names := map[string]string{
		"apt": "APT", "dnf": "DNF", "pacman": "Pacman", "zypper": "Zypper",
		"apk": "APK", "macos": "macOS", "homebrew": "Homebrew", "snap": "Snap",
		"npm": "npm", "flatpak": "Flatpak", "docker": "Docker",
		"openclaw": "OpenClaw", "distro": "Distro Release",
		"wordpress": "WordPress", "webproject": "Web Project",
	}
	if n, ok := names[typ]; ok {
		return n
	}
	return typ
}

func editAptWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	securityOnly := existing.GetBool("security_only", false)
	useSudo := existing.GetBool("use_sudo", true)
	hidePhased := existing.GetBool("hide_phased", false)

	err := huh.NewForm(
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
	if err != nil {
		return nil
	}

	existing.Options["security_only"] = securityOnly
	existing.Options["use_sudo"] = useSudo
	existing.Options["hide_phased"] = hidePhased
	fmt.Println("  APT watcher updated.")
	return nil
}

func editDnfWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	securityOnly := existing.GetBool("security_only", false)
	useSudo := existing.GetBool("use_sudo", true)

	err := huh.NewForm(
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
	if err != nil {
		return nil
	}

	existing.Options["security_only"] = securityOnly
	existing.Options["use_sudo"] = useSudo
	fmt.Println("  DNF watcher updated.")
	return nil
}

func editPacmanWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	useSudo := existing.GetBool("use_sudo", true)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use sudo for pacman sync operations?").
				Description(sudoDescription("pacman -Sy")).
				Value(&useSudo),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["use_sudo"] = useSudo
	fmt.Println("  Pacman watcher updated.")
	return nil
}

func editZypperWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	securityOnly := existing.GetBool("security_only", false)
	useSudo := existing.GetBool("use_sudo", true)

	err := huh.NewForm(
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
	if err != nil {
		return nil
	}

	existing.Options["security_only"] = securityOnly
	existing.Options["use_sudo"] = useSudo
	fmt.Println("  Zypper watcher updated.")
	return nil
}

func editApkWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	useSudo := existing.GetBool("use_sudo", false)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use sudo for apk operations?").
				Description(sudoDescription("apk update")).
				Value(&useSudo),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["use_sudo"] = useSudo
	fmt.Println("  APK watcher updated.")
	return nil
}

func editMacOSWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	securityOnly := existing.GetBool("security_only", false)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report security updates?").
				Value(&securityOnly),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["security_only"] = securityOnly
	fmt.Println("  macOS watcher updated.")
	return nil
}

func editHomebrewWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	includeCasks := existing.GetBool("include_casks", true)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include cask updates?").
				Description("Casks are macOS GUI applications (e.g. Firefox, VS Code)").
				Value(&includeCasks),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["include_casks"] = includeCasks
	fmt.Println("  Homebrew watcher updated.")
	return nil
}

func editDockerWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	containers := existing.GetString("containers", "all")
	excludeStr := strings.Join(existing.GetStringSlice("exclude", nil), ", ")

	err := huh.NewForm(
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
	if err != nil {
		return nil
	}

	existing.Options["containers"] = containers
	if excludeStr != "" {
		parts := strings.Split(excludeStr, ",")
		exclude := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				exclude = append(exclude, t)
			}
		}
		existing.Options["exclude"] = exclude
	} else {
		delete(existing.Options, "exclude")
	}
	fmt.Println("  Docker watcher updated.")
	return nil
}

func editOpenClawWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	channel := existing.GetString("channel", "")
	binaryPath := existing.GetString("binary_path", "")

	err := huh.NewForm(
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
			huh.NewInput().
				Title("Binary path (leave empty for auto-detect)").
				Value(&binaryPath),
		),
	).Run()
	if err != nil {
		return nil
	}

	if channel != "" {
		existing.Options["channel"] = channel
	} else {
		delete(existing.Options, "channel")
	}
	if binaryPath != "" {
		existing.Options["binary_path"] = binaryPath
	} else {
		delete(existing.Options, "binary_path")
	}
	fmt.Println("  OpenClaw watcher updated.")
	return nil
}

func editDistroWatcher(_ *config.Config, existing *config.WatcherConfig) error {
	ltsOnly := existing.GetBool("lts_only", true)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Only report LTS upgrades?").
				Description("Ubuntu only: Skip short-lived releases (e.g. 23.10) and only report upgrades to the next Long Term Support version (e.g. 22.04 → 24.04). Has no effect on Debian or Fedora.").
				Value(&ltsOnly),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["lts_only"] = ltsOnly
	fmt.Println("  Distro release watcher updated.")
	return nil
}

func editWordPressSite(cfg *config.Config, watcherIdx, siteIdx int) error {
	w := &cfg.Watchers[watcherIdx]
	sites := w.GetMapSlice("sites")
	if siteIdx < 0 || siteIdx >= len(sites) {
		return nil
	}
	site := sites[siteIdx]

	siteName, _ := site["name"].(string)
	sitePath, _ := site["path"].(string)
	envStr, _ := site["environment"].(string)
	runAs, _ := site["run_as"].(string)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Site name").
				Description("A human-readable name for this site").
				Value(&siteName),
			huh.NewInput().
				Title("WordPress path").
				Description("Full path to WordPress project root").
				Value(&sitePath),
		),
	).Run()
	if err != nil || sitePath == "" {
		return nil
	}

	// Environment selection
	detectedEnv := wordpress.DetectEnvironment(sitePath)
	if envStr == "" {
		envStr = string(detectedEnv)
	}

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
		return nil
	}

	selectedEnv := wordpress.Environment(envStr)

	if !selectedEnv.NeedsRunAs() {
		runAs = ""
	}

	// Update site in-place
	site["name"] = siteName
	site["path"] = sitePath
	site["environment"] = envStr
	if runAs != "" {
		site["run_as"] = runAs
	} else {
		delete(site, "run_as")
	}
	sites[siteIdx] = site

	// Write back to watcher options
	sitesIface := make([]interface{}, len(sites))
	for i, s := range sites {
		sitesIface[i] = s
	}
	w.Options["sites"] = sitesIface

	fmt.Printf("  WordPress site %q updated.\n", siteName)
	return nil
}

func editWebProjectEntry(cfg *config.Config, watcherIdx, projectIdx int) error {
	w := &cfg.Watchers[watcherIdx]
	projects := w.GetMapSlice("projects")
	if projectIdx < 0 || projectIdx >= len(projects) {
		return nil
	}
	project := projects[projectIdx]

	projectName, _ := project["name"].(string)
	projectPath, _ := project["path"].(string)
	envStr, _ := project["environment"].(string)
	checkAudit := true
	if v, ok := project["check_audit"].(bool); ok {
		checkAudit = v
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("A human-readable name for this project").
				Value(&projectName),
			huh.NewInput().
				Title("Project path").
				Description("Full path to project root").
				Value(&projectPath),
		),
	).Run()
	if err != nil || projectPath == "" {
		return nil
	}

	// Environment selection
	detectedEnv := webproject.DetectEnvironment(projectPath)
	if envStr == "" {
		envStr = string(detectedEnv)
	}

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
			huh.NewConfirm().
				Title("Run security audits?").
				Description("Check for known vulnerabilities (npm audit, composer audit, etc.)").
				Value(&checkAudit),
		),
	).Run()
	if err != nil {
		return nil
	}

	selectedEnv := webproject.Environment(envStr)

	// Run-as user: only kept if already configured, cleared for non-native envs
	runAs, _ := project["run_as"].(string)
	if !selectedEnv.NeedsRunAs() {
		runAs = ""
	}

	// Update project in-place
	project["name"] = projectName
	project["path"] = projectPath
	project["environment"] = envStr
	project["check_audit"] = checkAudit
	if runAs != "" {
		project["run_as"] = runAs
	} else {
		delete(project, "run_as")
	}
	projects[projectIdx] = project

	// Write back to watcher options
	projectsIface := make([]interface{}, len(projects))
	for i, p := range projects {
		projectsIface[i] = p
	}
	w.Options["projects"] = projectsIface

	fmt.Printf("  Web project %q updated.\n", projectName)
	return nil
}
