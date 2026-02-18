package apk

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
	checker.Register("apk", NewFromConfig)
}

// ApkChecker checks for available APK package updates (Alpine Linux).
type ApkChecker struct {
	useSudo bool
}

// NewFromConfig creates an ApkChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &ApkChecker{
		useSudo: cfg.GetBool("use_sudo", false),
	}, nil
}

func (a *ApkChecker) Name() string { return "apk" }

func (a *ApkChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: a.Name(),
		CheckedAt:   time.Now(),
	}

	// Update package index
	slog.Info("updating apk package index")
	var updateResult *executil.Result
	var err error

	updateResult, err = executil.RunMaybeSudo(a.useSudo, "apk", "update")
	if err != nil {
		slog.Warn("apk update failed, continuing with possibly stale data", "error", err, "stderr", updateResult.Stderr)
		result.Error = fmt.Sprintf("apk update failed: %s", err)
	}

	// Check for upgradable packages
	slog.Info("checking for upgradable packages")
	listResult, err := executil.Run("apk", "version", "-l", "<")
	if err != nil {
		return result, fmt.Errorf("apk version failed: %w", err)
	}

	result.Updates = parseVersionOutput(listResult.Stdout)

	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
