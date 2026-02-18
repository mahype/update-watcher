package snap

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
	checker.Register("snap", NewFromConfig)
}

// SnapChecker checks for available Snap package updates.
type SnapChecker struct{}

// NewFromConfig creates a SnapChecker from a watcher configuration.
func NewFromConfig(_ config.WatcherConfig) (checker.Checker, error) {
	return &SnapChecker{}, nil
}

func (s *SnapChecker) Name() string { return "snap" }

func (s *SnapChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: s.Name(),
		CheckedAt:   time.Now(),
	}

	slog.Info("checking for snap updates")
	listResult, err := executil.Run("snap", "refresh", "--list")
	if err != nil {
		// snap refresh --list exits 0 even when no updates;
		// an error here means snap itself failed.
		if listResult != nil && listResult.Stdout == "" {
			// "All snaps up to date." goes to stderr — no updates.
			result.Summary = "all snaps are up to date"
			return result, nil
		}
		return result, fmt.Errorf("snap refresh --list failed: %w", err)
	}

	result.Updates = parseRefreshList(listResult.Stdout)

	result.Summary = checker.BuildSummary(result.Updates, "snaps")

	return result, nil
}
