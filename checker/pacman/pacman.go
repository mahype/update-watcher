package pacman

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

func init() {
	checker.Register("pacman", NewFromConfig)
}

// PacmanChecker checks for available Pacman package updates (Arch/Manjaro).
type PacmanChecker struct {
	useSudo bool
}

// NewFromConfig creates a PacmanChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &PacmanChecker{
		useSudo: cfg.GetBool("use_sudo", true),
	}, nil
}

func (p *PacmanChecker) Name() string { return "pacman" }

func (p *PacmanChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: p.Name(),
		CheckedAt:   time.Now(),
	}

	// Sync package database
	slog.Info("syncing pacman package database")
	var syncResult *executil.Result
	var err error

	if p.useSudo {
		syncResult, err = executil.RunAsSudo("pacman", "-Sy")
	} else {
		syncResult, err = executil.Run("pacman", "-Sy")
	}
	if err != nil {
		slog.Warn("pacman -Sy failed, continuing with possibly stale data", "error", err, "stderr", syncResult.Stderr)
		result.Error = fmt.Sprintf("pacman -Sy failed: %s", err)
	}

	// List upgradable packages (no sudo needed)
	slog.Info("checking for upgradable packages")
	listResult, err := executil.Run("pacman", "-Qu")
	if err != nil {
		// Exit code 1 means no updates available
		if listResult != nil && listResult.ExitCode == 1 {
			result.Summary = "all packages are up to date"
			return result, nil
		}
		return result, fmt.Errorf("pacman -Qu failed: %w", err)
	}

	result.Updates = parseUpgradable(listResult.Stdout)

	if len(result.Updates) == 0 {
		result.Summary = "all packages are up to date"
	} else {
		result.Summary = fmt.Sprintf("%d packages", len(result.Updates))
	}

	return result, nil
}
