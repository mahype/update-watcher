package wordpress

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mahype/update-watcher/internal/executil"
)

// phpEnv suppresses PHP deprecation/warning/notice messages.
var phpEnv = []string{"WP_CLI_PHP_ARGS=-d error_reporting=E_ERROR -d display_errors=Off"}

// WPCLIRunner wraps WP-CLI command execution.
type WPCLIRunner struct {
	Path        string
	RunAs       string
	Environment Environment
	ProjectDir  string // Project root for container-based environments
}

// Run executes a wp-cli command with --format=json and returns the raw output.
func (w *WPCLIRunner) Run(args ...string) ([]byte, error) {
	fullArgs := append(args, "--format=json", "--quiet")
	result, err := w.exec(fullArgs)
	if err != nil {
		stderr := ""
		if result != nil {
			stderr = filterStderr(result.Stderr)
		}
		return nil, fmt.Errorf("wp-cli failed: %w (stderr: %s)", err, stderr)
	}
	return stripNonJSON([]byte(result.Stdout)), nil
}

// RunRaw executes a wp-cli command without --format=json flag.
func (w *WPCLIRunner) RunRaw(args ...string) (string, error) {
	fullArgs := append(args, "--quiet")
	result, err := w.exec(fullArgs)
	if err != nil {
		stderr := ""
		if result != nil {
			stderr = filterStderr(result.Stderr)
		}
		return "", fmt.Errorf("wp-cli failed: %w (stderr: %s)", err, stderr)
	}
	return strings.TrimSpace(result.Stdout), nil
}

// exec runs the WP-CLI command using the appropriate method for the environment.
func (w *WPCLIRunner) exec(wpArgs []string) (*executil.Result, error) {
	spec := BuildCommand(w.Environment, w.Path, w.ProjectDir)

	// Build the full argument list: spec.Args + wpArgs
	allArgs := make([]string, 0, len(spec.Args)+len(wpArgs))
	allArgs = append(allArgs, spec.Args...)
	allArgs = append(allArgs, wpArgs...)

	// Container-based environments: run in project directory, no sudo
	if w.Environment.IsContainerBased() {
		return executil.RunInDirWithEnv(spec.WorkDir, spec.Env, spec.Command, allArgs...)
	}

	// Native environment with run_as: use sudo -u
	if spec.NeedsRunAs && w.RunAs != "" {
		return executil.RunAsUserWithEnv(spec.Env, w.RunAs, spec.Command, allArgs...)
	}

	// Host-based environments (MAMP, XAMPP, Valet, Bedrock, etc.)
	if len(spec.Env) > 0 {
		return executil.RunWithEnv(spec.Env, spec.Command, allArgs...)
	}

	return executil.Run(spec.Command, allArgs...)
}

// stripNonJSON removes any PHP warnings/notices that appear before the JSON output.
func stripNonJSON(output []byte) []byte {
	for i, b := range output {
		if b == '[' || b == '{' {
			return output[i:]
		}
	}
	return output
}

// filterStderr removes PHP warning/deprecation/notice lines from stderr.
func filterStderr(stderr string) string {
	var filtered []string
	for _, line := range strings.Split(stderr, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "PHP Deprecated:") ||
			strings.HasPrefix(trimmed, "PHP Warning:") ||
			strings.HasPrefix(trimmed, "PHP Notice:") ||
			strings.HasPrefix(trimmed, "Deprecated:") ||
			strings.HasPrefix(trimmed, "Warning:") ||
			strings.HasPrefix(trimmed, "Notice:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// CoreUpdate represents a WP-CLI core check-update JSON entry.
type CoreUpdate struct {
	Version string `json:"version"`
	Update  string `json:"update_type"` // "major", "minor", "patch"
}

// PluginInfo represents a WP-CLI plugin list JSON entry.
type PluginInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Update    string `json:"update"`
	Version   string `json:"version"`
	UpdateVer string `json:"update_version"`
}

// ThemeInfo represents a WP-CLI theme list JSON entry.
type ThemeInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Update    string `json:"update"`
	Version   string `json:"version"`
	UpdateVer string `json:"update_version"`
}

// CheckCoreUpdates returns available core updates.
func (w *WPCLIRunner) CheckCoreUpdates() ([]CoreUpdate, string, error) {
	// Get current version using RunRaw (core version doesn't support --format=json)
	currentVersion, err := w.RunRaw("core", "version")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get core version: %w", err)
	}

	output, err := w.Run("core", "check-update")
	if err != nil {
		return nil, currentVersion, nil // No updates available is not an error
	}

	var updates []CoreUpdate
	if err := json.Unmarshal(output, &updates); err != nil {
		return nil, currentVersion, nil // Parse failures mean no structured updates
	}

	return updates, currentVersion, nil
}

// CheckPluginUpdates returns plugins with available updates.
func (w *WPCLIRunner) CheckPluginUpdates() ([]PluginInfo, error) {
	output, err := w.Run("plugin", "list", "--update=available")
	if err != nil {
		return nil, err
	}

	var plugins []PluginInfo
	if err := json.Unmarshal(output, &plugins); err != nil {
		return nil, fmt.Errorf("failed to parse plugin list: %w", err)
	}

	return plugins, nil
}

// CheckThemeUpdates returns themes with available updates.
func (w *WPCLIRunner) CheckThemeUpdates() ([]ThemeInfo, error) {
	output, err := w.Run("theme", "list", "--update=available")
	if err != nil {
		return nil, err
	}

	var themes []ThemeInfo
	if err := json.Unmarshal(output, &themes); err != nil {
		return nil, fmt.Errorf("failed to parse theme list: %w", err)
	}

	return themes, nil
}
