package config

import (
	"fmt"
	"strings"
)

// ValidationError collects multiple validation issues.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation failed:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

func (e *ValidationError) add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) hasErrors() bool {
	return len(e.Errors) > 0
}

// Validate checks the configuration for common issues.
func Validate(cfg *Config) error {
	ve := &ValidationError{}

	if len(cfg.Watchers) == 0 {
		ve.add("no watchers configured")
	}

	validTypes := map[string]bool{"apt": true, "dnf": true, "pacman": true, "zypper": true, "apk": true, "docker": true, "wordpress": true, "webproject": true, "macos": true}
	for i, w := range cfg.Watchers {
		if !validTypes[w.Type] {
			ve.add(fmt.Sprintf("watcher[%d]: unknown type %q", i, w.Type))
		}
		if w.Type == "wordpress" {
			sites := w.GetMapSlice("sites")
			if len(sites) == 0 {
				ve.add(fmt.Sprintf("watcher[%d] (wordpress): no sites configured", i))
			}
			for j, site := range sites {
				if _, ok := site["path"]; !ok {
					ve.add(fmt.Sprintf("watcher[%d] (wordpress) site[%d]: missing path", i, j))
				}
			}
		}
		if w.Type == "webproject" {
			projects := w.GetMapSlice("projects")
			if len(projects) == 0 {
				ve.add(fmt.Sprintf("watcher[%d] (webproject): no projects configured", i))
			}
			for j, project := range projects {
				if _, ok := project["path"]; !ok {
					ve.add(fmt.Sprintf("watcher[%d] (webproject) project[%d]: missing path", i, j))
				}
			}
		}
	}

	if len(cfg.Notifiers) == 0 {
		ve.add("no notifiers configured")
	}

	validNotifiers := map[string]bool{
		"slack":    true,
		"ntfy":     true,
		"webhook":  true,
		"discord":  true,
		"telegram": true,
		"teams":    true,
		"email":    true,
	}
	for i, n := range cfg.Notifiers {
		if !validNotifiers[n.Type] {
			ve.add(fmt.Sprintf("notifier[%d]: unknown type %q", i, n.Type))
		}

		if !n.Enabled {
			continue
		}

		opts := WatcherConfig{Options: n.Options}

		switch n.Type {
		case "slack":
			if opts.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (slack): missing webhook_url", i))
			}
		case "ntfy":
			if opts.GetString("topic", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (ntfy): missing topic", i))
			}
		case "webhook":
			if opts.GetString("url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (webhook): missing url", i))
			}
		case "discord":
			if opts.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (discord): missing webhook_url", i))
			}
		case "telegram":
			if opts.GetString("bot_token", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (telegram): missing bot_token", i))
			}
			if opts.GetString("chat_id", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (telegram): missing chat_id", i))
			}
		case "teams":
			if opts.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (teams): missing webhook_url", i))
			}
		case "email":
			for _, field := range []string{"smtp_host", "username", "password", "from"} {
				if opts.GetString(field, "") == "" {
					ve.add(fmt.Sprintf("notifier[%d] (email): missing %s", i, field))
				}
			}
			to := opts.GetStringSlice("to", nil)
			if len(to) == 0 {
				ve.add(fmt.Sprintf("notifier[%d] (email): missing 'to' recipients", i))
			}
		}
	}

	policy := cfg.Settings.SendPolicy
	if policy != "always" && policy != "only-on-updates" {
		ve.add(fmt.Sprintf("settings.send_policy: invalid value %q (must be \"always\" or \"only-on-updates\")", policy))
	}

	if ve.hasErrors() {
		return ve
	}
	return nil
}
