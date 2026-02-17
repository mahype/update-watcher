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

	validTypes := map[string]bool{"apt": true, "docker": true, "wordpress": true}
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
	}

	if len(cfg.Notifiers) == 0 {
		ve.add("no notifiers configured")
	}

	validNotifiers := map[string]bool{"slack": true}
	for i, n := range cfg.Notifiers {
		if !validNotifiers[n.Type] {
			ve.add(fmt.Sprintf("notifier[%d]: unknown type %q", i, n.Type))
		}
		if n.Type == "slack" && n.Enabled {
			opts := WatcherConfig{Options: n.Options}
			url := opts.GetString("webhook_url", "")
			if url == "" {
				ve.add(fmt.Sprintf("notifier[%d] (slack): missing webhook_url", i))
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
