package apt

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

func init() {
	checker.Register("apt", NewFromConfig)
}

// AptChecker checks for available APT package updates.
type AptChecker struct {
	useSudo      bool
	securityOnly bool
}

// NewFromConfig creates an AptChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &AptChecker{
		useSudo:      cfg.GetBool("use_sudo", true),
		securityOnly: cfg.GetBool("security_only", false),
	}, nil
}

func (a *AptChecker) Name() string { return "apt" }

func (a *AptChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: a.Name(),
		CheckedAt:   time.Now(),
	}

	// Refresh package lists
	slog.Info("refreshing apt package lists")
	var refreshResult *executil.Result
	var err error

	if a.useSudo {
		refreshResult, err = executil.RunAsSudo("apt-get", "update", "-qq")
	} else {
		refreshResult, err = executil.Run("apt-get", "update", "-qq")
	}
	if err != nil {
		slog.Warn("apt-get update failed, continuing with possibly stale data", "error", err, "stderr", refreshResult.Stderr)
		result.Error = fmt.Sprintf("apt-get update failed: %s", err)
	}

	// Get upgradable packages
	slog.Info("checking for upgradable packages")
	listResult, err := executil.Run("apt", "list", "--upgradable")
	if err != nil {
		return result, fmt.Errorf("apt list --upgradable failed: %w", err)
	}

	result.Updates = parseUpgradable(listResult.Stdout, a.securityOnly)

	// Build summary
	secCount := 0
	for _, u := range result.Updates {
		if u.Type == checker.UpdateTypeSecurity {
			secCount++
		}
	}

	if len(result.Updates) == 0 {
		result.Summary = "all packages are up to date"
	} else if secCount > 0 {
		result.Summary = fmt.Sprintf("%d packages (%d security)", len(result.Updates), secCount)
	} else {
		result.Summary = fmt.Sprintf("%d packages", len(result.Updates))
	}

	return result, nil
}
