package apt

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
	checker.Register("apt", NewFromConfig)
}

// AptChecker checks for available APT package updates.
type AptChecker struct {
	useSudo      bool
	securityOnly bool
	hidePhased   bool
}

// NewFromConfig creates an AptChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &AptChecker{
		useSudo:      cfg.GetBool("use_sudo", true),
		securityOnly: cfg.GetBool("security_only", false),
		hidePhased:   cfg.GetBool("hide_phased", false),
	}, nil
}

func (a *AptChecker) Name() string { return "apt" }

func (a *AptChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: a.Name(),
		CheckedAt:   time.Now(),
	}

	// Refresh package lists
	slog.Info("refreshing apt package lists")
	var refreshResult *executil.Result
	var err error

	refreshResult, err = executil.RunMaybeSudo(a.useSudo, "apt-get", "update")
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

	// Detect phased updates via dry-run simulation.
	// apt list --upgradable does not always show the [phased X%] marker,
	// but apt-get -s upgrade reliably reports "deferred due to phasing".
	slog.Info("detecting phased updates via dry-run")
	simResult, err := executil.Run("apt-get", "-s", "upgrade")
	if err != nil {
		slog.Warn("apt-get -s upgrade failed, skipping phased detection", "error", err)
	} else {
		deferred := parseDeferredPackages(simResult.Stdout)
		for i := range result.Updates {
			if result.Updates[i].Phasing == "" && deferred[result.Updates[i].Name] {
				result.Updates[i].Phasing = "deferred"
			}
		}
	}

	// Filter out phased updates if configured.
	if a.hidePhased {
		filtered := result.Updates[:0]
		for _, u := range result.Updates {
			if u.Phasing == "" {
				filtered = append(filtered, u)
			}
		}
		result.Updates = filtered
	}

	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
