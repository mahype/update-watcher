// Package sudoers generates and manages the /etc/sudoers.d/update-watcher
// file based on the watcher configuration. It grants the dedicated service
// user NOPASSWD access only to the specific commands each configured checker
// actually invokes.
package sudoers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
	"github.com/mahype/update-watcher/internal/rootcheck"
)

const (
	// TargetPath is where the generated sudoers file is placed.
	TargetPath = "/etc/sudoers.d/update-watcher"

	fileMode = 0o440
)

// Rule is a single NOPASSWD sudoers entry.
type Rule struct {
	RunAs   string   // target user, e.g. "root" or "www-data"
	Command string   // absolute binary path, e.g. "/usr/bin/apt-get"
	Args    []string // optional argument restriction; nil/empty = any args
}

// LookPathFunc resolves a binary name to its absolute path. Injected so tests
// can stub it out without relying on the host's PATH.
type LookPathFunc func(string) (string, error)

var defaultLookPath LookPathFunc = exec.LookPath

// Build derives the sudoers rules from the configuration. Warnings describe
// watchers that are configured but whose binary cannot be resolved on this
// host — we skip the rule rather than fail the install.
func Build(cfg *config.Config) ([]Rule, []string, error) {
	return BuildWith(cfg, defaultLookPath)
}

// BuildWith is Build with an injectable LookPath.
func BuildWith(cfg *config.Config, lookPath LookPathFunc) ([]Rule, []string, error) {
	if cfg == nil {
		return nil, nil, nil
	}

	var rules []Rule
	var warnings []string

	for _, w := range cfg.Watchers {
		if !w.Enabled {
			continue
		}
		switch w.Type {
		case "apt":
			if !w.GetBool("use_sudo", true) {
				continue
			}
			path, err := lookPath("apt-get")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("apt watcher has use_sudo but 'apt-get' not found in PATH: %v", err))
				continue
			}
			rules = append(rules, Rule{RunAs: "root", Command: path, Args: []string{"update"}})

		case "dnf":
			if !w.GetBool("use_sudo", true) {
				continue
			}
			path, err := lookPath("dnf")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("dnf watcher has use_sudo but 'dnf' not found in PATH: %v", err))
				continue
			}
			rules = append(rules,
				Rule{RunAs: "root", Command: path, Args: []string{"updateinfo", "list", "--security"}},
				Rule{RunAs: "root", Command: path, Args: []string{"check-update"}},
			)

		case "zypper":
			if !w.GetBool("use_sudo", true) {
				continue
			}
			path, err := lookPath("zypper")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("zypper watcher has use_sudo but 'zypper' not found in PATH: %v", err))
				continue
			}
			rules = append(rules,
				Rule{RunAs: "root", Command: path, Args: []string{"--non-interactive", "refresh"}},
				Rule{RunAs: "root", Command: path, Args: []string{"--non-interactive", "list-patches", "--category", "security"}},
				Rule{RunAs: "root", Command: path, Args: []string{"--non-interactive", "list-updates"}},
			)

		case "pacman":
			if !w.GetBool("use_sudo", true) {
				continue
			}
			path, err := lookPath("pacman")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("pacman watcher has use_sudo but 'pacman' not found in PATH: %v", err))
				continue
			}
			rules = append(rules, Rule{RunAs: "root", Command: path, Args: []string{"-Sy"}})

		case "apk":
			if !w.GetBool("use_sudo", false) {
				continue
			}
			path, err := lookPath("apk")
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("apk watcher has use_sudo but 'apk' not found in PATH: %v", err))
				continue
			}
			rules = append(rules, Rule{RunAs: "root", Command: path, Args: []string{"update"}})

		case "wordpress":
			wpRules, warn := wordpressRules(w, lookPath)
			rules = append(rules, wpRules...)
			if warn != "" {
				warnings = append(warnings, warn)
			}
		}
	}

	return dedup(rules), warnings, nil
}

// wordpressRules emits one NOPASSWD rule per unique run_as user in the
// wordpress watcher. WP-CLI is granted unrestricted args because its
// subcommands (core/plugin/theme × flags) are too numerous to enumerate
// precisely and the risk is bounded by the run_as user (e.g. www-data).
func wordpressRules(w config.WatcherConfig, lookPath LookPathFunc) ([]Rule, string) {
	sites := w.GetMapSlice("sites")
	if len(sites) == 0 {
		return nil, ""
	}

	var rules []Rule
	seen := make(map[string]bool)
	var wpPath string
	needsLookup := false

	for _, s := range sites {
		runAs, _ := s["run_as"].(string)
		if runAs == "" || seen[runAs] {
			continue
		}
		needsLookup = true
		break
	}
	if !needsLookup {
		return nil, ""
	}

	path, err := lookPath("wp")
	if err != nil {
		return nil, fmt.Sprintf("wordpress watcher has sites with run_as but 'wp' (wp-cli) not found in PATH: %v", err)
	}
	wpPath = path

	for _, s := range sites {
		runAs, _ := s["run_as"].(string)
		if runAs == "" || seen[runAs] {
			continue
		}
		rules = append(rules, Rule{RunAs: runAs, Command: wpPath})
		seen[runAs] = true
	}
	return rules, ""
}

func dedup(rules []Rule) []Rule {
	seen := make(map[string]bool)
	out := make([]Rule, 0, len(rules))
	for _, r := range rules {
		key := r.RunAs + "|" + r.Command + "|" + strings.Join(r.Args, " ")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

// Render formats the rules into the full sudoers file content.
func Render(rules []Rule, serviceUser string) string {
	sorted := make([]Rule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].RunAs != sorted[j].RunAs {
			return sorted[i].RunAs < sorted[j].RunAs
		}
		if sorted[i].Command != sorted[j].Command {
			return sorted[i].Command < sorted[j].Command
		}
		return strings.Join(sorted[i].Args, " ") < strings.Join(sorted[j].Args, " ")
	})

	var b strings.Builder
	b.WriteString("# Managed by update-watcher — do not edit manually.\n")
	b.WriteString("# Regenerate with: update-watcher install-cron\n")
	b.WriteString("# Remove with:     update-watcher uninstall-cron\n")
	b.WriteString("#\n")
	b.WriteString("# This file grants NOPASSWD access to specific commands required by\n")
	b.WriteString("# configured watchers. It is rewritten from scratch on every install-cron run.\n\n")
	for _, r := range sorted {
		cmdLine := r.Command
		if len(r.Args) > 0 {
			cmdLine = r.Command + " " + strings.Join(r.Args, " ")
		}
		fmt.Fprintf(&b, "%s ALL=(%s) NOPASSWD: %s\n", serviceUser, r.RunAs, cmdLine)
	}
	return b.String()
}

// Write atomically writes the sudoers file after validating it with visudo.
// Caller must run as root; we do not sudo ourselves.
func Write(rules []Rule) error {
	content := Render(rules, rootcheck.ServiceUserName())

	dir := filepath.Dir(TargetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create sudoers.d dir: %w", err)
	}

	tmp := TargetPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), fileMode); err != nil {
		return fmt.Errorf("write temp sudoers file: %w", err)
	}

	res, err := executil.Run("visudo", "-cf", tmp)
	if err != nil {
		_ = os.Remove(tmp)
		detail := ""
		if res != nil {
			detail = strings.TrimSpace(res.Stderr)
			if detail == "" {
				detail = strings.TrimSpace(res.Stdout)
			}
		}
		if detail != "" {
			return fmt.Errorf("visudo validation failed: %w (%s)", err, detail)
		}
		return fmt.Errorf("visudo validation failed: %w", err)
	}

	if err := os.Rename(tmp, TargetPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename sudoers file: %w", err)
	}
	if err := os.Chmod(TargetPath, fileMode); err != nil {
		return fmt.Errorf("chmod sudoers file: %w", err)
	}
	return nil
}

// Remove deletes the sudoers file. Missing file is not an error.
func Remove() error {
	if err := os.Remove(TargetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove sudoers file: %w", err)
	}
	return nil
}

// FormatRule returns the sudoers line for a rule, for user-facing output.
func FormatRule(serviceUser string, r Rule) string {
	cmdLine := r.Command
	if len(r.Args) > 0 {
		cmdLine = r.Command + " " + strings.Join(r.Args, " ")
	}
	return fmt.Sprintf("%s ALL=(%s) NOPASSWD: %s", serviceUser, r.RunAs, cmdLine)
}
