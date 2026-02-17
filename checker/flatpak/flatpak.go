package flatpak

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

func init() {
	checker.Register("flatpak", NewFromConfig)
}

// FlatpakChecker checks for available Flatpak application updates.
type FlatpakChecker struct{}

// NewFromConfig creates a FlatpakChecker from a watcher configuration.
func NewFromConfig(_ config.WatcherConfig) (checker.Checker, error) {
	return &FlatpakChecker{}, nil
}

func (f *FlatpakChecker) Name() string { return "flatpak" }

func (f *FlatpakChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: f.Name(),
		CheckedAt:   time.Now(),
	}

	slog.Info("checking for flatpak updates")
	listResult, err := executil.Run("flatpak", "remote-ls", "--updates", "--app", "--columns=name,application,version")
	if err != nil {
		return result, fmt.Errorf("flatpak remote-ls failed: %w", err)
	}

	result.Updates = parseRemoteUpdates(listResult.Stdout)

	if len(result.Updates) == 0 {
		result.Summary = "all flatpaks are up to date"
	} else {
		result.Summary = fmt.Sprintf("%d applications", len(result.Updates))
	}

	return result, nil
}
