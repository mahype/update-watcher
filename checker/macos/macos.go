package macos

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

func init() {
	checker.Register("macos", NewFromConfig)
}

// MacOSChecker checks for available macOS software updates.
type MacOSChecker struct {
	securityOnly bool
}

// NewFromConfig creates a MacOSChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &MacOSChecker{
		securityOnly: cfg.GetBool("security_only", false),
	}, nil
}

func (m *MacOSChecker) Name() string { return "macos" }

func (m *MacOSChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: m.Name(),
		CheckedAt:   time.Now(),
	}

	slog.Info("checking for macOS software updates")
	listResult, err := executil.Run("softwareupdate", "-l")
	if err != nil {
		return result, fmt.Errorf("softwareupdate -l failed: %w", err)
	}

	result.Updates = parseSoftwareUpdate(listResult.Stdout, m.securityOnly)

	// Build summary
	secCount := 0
	for _, u := range result.Updates {
		if u.Type == checker.UpdateTypeSecurity {
			secCount++
		}
	}

	if len(result.Updates) == 0 {
		result.Summary = "all software is up to date"
	} else if secCount > 0 {
		result.Summary = fmt.Sprintf("%d updates (%d security)", len(result.Updates), secCount)
	} else {
		result.Summary = fmt.Sprintf("%d updates", len(result.Updates))
	}

	return result, nil
}
