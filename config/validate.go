package config

import (
	"fmt"
	"strings"
)

// secretKeys are option keys that typically contain sensitive credentials.
var secretKeys = map[string]bool{
	"webhook_url":  true,
	"bot_token":    true,
	"password":     true,
	"access_token": true,
	"auth_header":  true,
	"routing_key":  true,
	"user_key":     true,
	"app_token":    true,
	"token":        true,
}

// WarnPlaintextSecrets checks for credential fields that contain literal values
// instead of ${ENV_VAR} references. Returns non-blocking warning messages.
func WarnPlaintextSecrets(cfg *Config) []string {
	var warnings []string
	for i, n := range cfg.Notifiers {
		if !n.Enabled {
			continue
		}
		for key := range secretKeys {
			val := n.Options.GetString(key, "")
			if val != "" && !strings.HasPrefix(val, "${") {
				warnings = append(warnings, fmt.Sprintf("notifiers[%d] (%s): %s contains a plaintext secret — consider using ${ENV_VAR} syntax", i, n.Type, key))
			}
		}
	}
	return warnings
}

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

	validTypes := map[string]bool{"apt": true, "dnf": true, "pacman": true, "zypper": true, "apk": true, "macos": true, "homebrew": true, "snap": true, "flatpak": true, "docker": true, "wordpress": true, "webproject": true, "openclaw": true, "distro": true}
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
		"slack":      true,
		"ntfy":       true,
		"webhook":    true,
		"discord":    true,
		"telegram":   true,
		"teams":      true,
		"email":      true,
		"pushover":   true,
		"gotify":     true,
		"googlechat": true,
		"matrix":     true,
		"mattermost": true,
		"rocketchat": true,
		"pagerduty":  true,
		"pushbullet": true,
	}
	for i, n := range cfg.Notifiers {
		if !validNotifiers[n.Type] {
			ve.add(fmt.Sprintf("notifier[%d]: unknown type %q", i, n.Type))
		}

		if !n.Enabled {
			continue
		}

		switch n.Type {
		case "slack":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (slack): missing webhook_url", i))
			}
		case "ntfy":
			if n.Options.GetString("topic", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (ntfy): missing topic", i))
			}
		case "webhook":
			if n.Options.GetString("url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (webhook): missing url", i))
			}
		case "discord":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (discord): missing webhook_url", i))
			}
		case "telegram":
			if n.Options.GetString("bot_token", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (telegram): missing bot_token", i))
			}
			if n.Options.GetString("chat_id", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (telegram): missing chat_id", i))
			}
		case "teams":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (teams): missing webhook_url", i))
			}
		case "email":
			for _, field := range []string{"smtp_host", "username", "password", "from"} {
				if n.Options.GetString(field, "") == "" {
					ve.add(fmt.Sprintf("notifier[%d] (email): missing %s", i, field))
				}
			}
			to := n.Options.GetStringSlice("to", nil)
			if len(to) == 0 {
				ve.add(fmt.Sprintf("notifier[%d] (email): missing 'to' recipients", i))
			}
		case "pushover":
			if n.Options.GetString("app_token", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (pushover): missing app_token", i))
			}
			if n.Options.GetString("user_key", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (pushover): missing user_key", i))
			}
		case "gotify":
			if n.Options.GetString("server_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (gotify): missing server_url", i))
			}
			if n.Options.GetString("token", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (gotify): missing token", i))
			}
		case "googlechat":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (googlechat): missing webhook_url", i))
			}
		case "matrix":
			for _, field := range []string{"homeserver", "access_token", "room_id"} {
				if n.Options.GetString(field, "") == "" {
					ve.add(fmt.Sprintf("notifier[%d] (matrix): missing %s", i, field))
				}
			}
		case "mattermost":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (mattermost): missing webhook_url", i))
			}
		case "rocketchat":
			if n.Options.GetString("webhook_url", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (rocketchat): missing webhook_url", i))
			}
		case "pagerduty":
			if n.Options.GetString("routing_key", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (pagerduty): missing routing_key", i))
			}
		case "pushbullet":
			if n.Options.GetString("access_token", "") == "" {
				ve.add(fmt.Sprintf("notifier[%d] (pushbullet): missing access_token", i))
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
