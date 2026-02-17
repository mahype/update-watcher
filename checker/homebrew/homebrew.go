package homebrew

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

func init() {
	checker.Register("homebrew", NewFromConfig)
}

// HomebrewChecker checks for available Homebrew package updates.
type HomebrewChecker struct {
	includeCasks bool
}

// NewFromConfig creates a HomebrewChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &HomebrewChecker{
		includeCasks: cfg.GetBool("include_casks", true),
	}, nil
}

func (h *HomebrewChecker) Name() string { return "homebrew" }

func (h *HomebrewChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: h.Name(),
		CheckedAt:   time.Now(),
	}

	// Update Homebrew database
	slog.Info("updating homebrew")
	updateResult, err := executil.Run("brew", "update")
	if err != nil {
		slog.Warn("brew update failed, continuing with possibly stale data", "error", err, "stderr", updateResult.Stderr)
		result.Error = fmt.Sprintf("brew update failed: %s", err)
	}

	// List outdated packages as JSON
	slog.Info("checking for outdated homebrew packages")
	listResult, err := executil.Run("brew", "outdated", "--json=v2")
	if err != nil {
		return result, fmt.Errorf("brew outdated failed: %w", err)
	}

	updates, err := parseOutdated(listResult.Stdout, h.includeCasks)
	if err != nil {
		return result, fmt.Errorf("failed to parse brew outdated output: %w", err)
	}
	result.Updates = updates

	// Build summary
	formulaeCount := 0
	caskCount := 0
	for _, u := range result.Updates {
		if u.Source == "casks" {
			caskCount++
		} else {
			formulaeCount++
		}
	}

	if len(result.Updates) == 0 {
		result.Summary = "all packages are up to date"
	} else if caskCount > 0 {
		result.Summary = fmt.Sprintf("%d formulae, %d casks", formulaeCount, caskCount)
	} else {
		result.Summary = fmt.Sprintf("%d formulae", formulaeCount)
	}

	return result, nil
}
