package output

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/cron"
)

// PrintStatus prints the current configuration as a status table.
func PrintStatus(cfg *config.Config, cronJobs []cron.InstalledJob) {
	fmt.Printf("\n=== Update Watcher Status ===\n\n")
	fmt.Printf("  Hostname:    %s\n", cfg.Hostname)
	fmt.Printf("  Config:      %s\n", config.ConfigPath())
	fmt.Printf("  Send Policy: %s\n", cfg.Settings.SendPolicy)
	if cfg.Settings.MinPriority != "" {
		fmt.Printf("  Min Priority: %s\n", cfg.Settings.MinPriority)
	}

	if len(cronJobs) == 0 {
		fmt.Printf("  Cron:        not installed\n")
	} else {
		for i, j := range cronJobs {
			prefix := "  Cron:        "
			if i > 0 {
				prefix = "               "
			}
			fmt.Printf("%s%s (%s)\n", prefix, cron.JobTypeLabel(j.Type), cron.FormatSchedule(j.Schedule))
		}
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
		if w.Type == "webproject" {
			projects := w.GetMapSlice("projects")
			detail = fmt.Sprintf(" (%d projects)", len(projects))
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
		detail := ""
		if n.SendPolicy != "" {
			detail += fmt.Sprintf(", policy: %s", n.SendPolicy)
		}
		if n.MinPriority != "" {
			detail += fmt.Sprintf(", min: %s+", n.MinPriority)
		}
		host := notifierHost(n)
		if host != "" {
			fmt.Printf("    - %-12s %-25s %s%s\n", n.Type, host, status, detail)
		} else {
			fmt.Printf("    - %-12s %s%s\n", n.Type, status, detail)
		}
	}

	fmt.Println()
}

// notifierHost extracts a host or identifier from the notifier's options
// so that multiple notifiers of the same type can be distinguished.
func notifierHost(n config.NotifierConfig) string {
	switch n.Type {
	case "slack", "discord", "teams", "googlechat", "mattermost", "rocketchat":
		return extractHost(n.Options.GetString("webhook_url", ""))
	case "webhook", "updatewall", "homeassistant":
		return extractHost(n.Options.GetString("url", ""))
	case "gotify":
		return extractHost(n.Options.GetString("server_url", ""))
	case "ntfy":
		if h := extractHost(n.Options.GetString("server_url", "")); h != "" {
			return h
		}
		return n.Options.GetString("topic", "")
	case "matrix":
		return extractHost(n.Options.GetString("homeserver", ""))
	case "email":
		return n.Options.GetString("smtp_host", "")
	case "telegram":
		return n.Options.GetString("chat_id", "")
	}
	return ""
}

// extractHost parses a URL and returns its hostname, or empty string on failure.
func extractHost(rawURL string) string {
	if rawURL == "" || strings.HasPrefix(rawURL, "${") {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
