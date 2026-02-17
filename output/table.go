package output

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
)

// PrintStatus prints the current configuration as a status table.
func PrintStatus(cfg *config.Config, cronInstalled bool, cronSchedule string) {
	fmt.Printf("\n=== Update Watcher Status ===\n\n")
	fmt.Printf("  Hostname:    %s\n", cfg.Hostname)
	fmt.Printf("  Config:      %s\n", config.ConfigPath())
	fmt.Printf("  Send Policy: %s\n", cfg.Settings.SendPolicy)

	if cronInstalled {
		fmt.Printf("  Cron:        installed (%s)\n", cronSchedule)
	} else {
		fmt.Printf("  Cron:        not installed\n")
	}

	fmt.Printf("\n  Watchers:\n")
	if len(cfg.Watchers) == 0 {
		fmt.Printf("    (none configured)\n")
	}
	for _, w := range cfg.Watchers {
		status := "enabled"
		if !w.Enabled {
			status = "disabled"
		}
		detail := ""
		if w.Type == "wordpress" {
			sites := w.GetMapSlice("sites")
			detail = fmt.Sprintf(" (%d sites)", len(sites))
		}
		if w.Type == "docker" {
			containers := w.GetString("containers", "all")
			detail = fmt.Sprintf(" (containers: %s)", containers)
		}
		fmt.Printf("    - %-12s %s%s\n", w.Type, status, detail)
	}

	fmt.Printf("\n  Notifiers:\n")
	if len(cfg.Notifiers) == 0 {
		fmt.Printf("    (none configured)\n")
	}
	for _, n := range cfg.Notifiers {
		status := "enabled"
		if !n.Enabled {
			status = "disabled"
		}
		fmt.Printf("    - %-12s %s\n", n.Type, status)
	}

	fmt.Println()
}
