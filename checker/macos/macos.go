package macos

import (
	"context"
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

func (m *MacOSChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
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

	result.Summary = checker.BuildSummary(result.Updates, "updates")

	return result, nil
}
