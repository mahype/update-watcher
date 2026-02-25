package openclaw

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
	checker.Register("openclaw", NewFromConfig)
}

// OpenClawChecker checks for available OpenClaw updates.
type OpenClawChecker struct {
	channel    string
	binaryPath string
}

// NewFromConfig creates an OpenClawChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &OpenClawChecker{
		channel:    cfg.GetString("channel", ""),
		binaryPath: cfg.GetString("binary_path", "openclaw"),
	}, nil
}

func (o *OpenClawChecker) Name() string { return "openclaw" }

func (o *OpenClawChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: o.Name(),
		CheckedAt:   time.Now(),
	}

	// Get current version.
	slog.Info("checking openclaw version")
	currentVersion := ""
	verResult, err := executil.Run(o.binaryPath, "--version")
	if err == nil {
		currentVersion = parseVersion(verResult.Stdout)
	} else {
		slog.Warn("could not determine current openclaw version", "error", err)
	}

	// Check for updates.
	slog.Info("checking for openclaw updates")
	args := []string{"update", "status"}
	if o.channel != "" {
		args = append(args, "--channel", o.channel)
	}
	statusResult, err := executil.RunWithTimeout(30*time.Second, o.binaryPath, args...)
	if err != nil {
		return result, fmt.Errorf("openclaw update status failed: %w", err)
	}

	newVersion, available := parseStatus(statusResult.Stdout)
	if !available {
		result.Summary = checker.BuildSummary(nil, "components")
		return result, nil
	}

	result.Updates = []checker.Update{
		{
			Name:           "openclaw",
			CurrentVersion: currentVersion,
			NewVersion:     newVersion,
			Type:           checker.UpdateTypeCore,
			Priority:       checker.PriorityNormal,
		},
	}
	result.Summary = checker.BuildSummary(result.Updates, "components")

	return result, nil
}
