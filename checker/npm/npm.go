package npm

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
	checker.Register("npm", NewFromConfig)
}

// NpmChecker checks for outdated globally installed npm packages.
type NpmChecker struct{}

// NewFromConfig creates an NpmChecker from a watcher configuration.
func NewFromConfig(_ config.WatcherConfig) (checker.Checker, error) {
	return &NpmChecker{}, nil
}

func (n *NpmChecker) Name() string { return "npm" }

func (n *NpmChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: n.Name(),
		CheckedAt:   time.Now(),
	}

	slog.Info("checking for outdated global npm packages")
	listResult, err := executil.Run("npm", "outdated", "-g", "--json")
	// npm outdated exits with code 1 when outdated packages exist.
	if err != nil && listResult == nil {
		return result, fmt.Errorf("npm outdated -g failed: %w", err)
	}

	updates, err := parseOutdated(listResult.Stdout)
	if err != nil {
		return result, fmt.Errorf("failed to parse npm outdated output: %w", err)
	}
	result.Updates = updates
	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
