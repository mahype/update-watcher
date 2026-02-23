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
		hidePhased:   cfg.GetBool("hide_phased", true),
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

	// Use dry-run simulation for phased update detection and security cross-check.
	// apt list --upgradable does not always show the [phased X%] marker,
	// but apt-get -s dist-upgrade reliably reports "deferred due to phasing".
	// It also provides Inst lines with full origin info for security classification.
	slog.Info("detecting phased updates and security origins via dry-run")
	simResult, err := executil.Run("apt-get", "-s", "dist-upgrade")
	if err != nil {
		slog.Warn("apt-get -s dist-upgrade failed, skipping phased/security detection", "error", err)
	} else {
		deferred := parseDeferredPackages(simResult.Stdout)
		for i := range result.Updates {
			if result.Updates[i].Phasing == "" && deferred[result.Updates[i].Name] {
				result.Updates[i].Phasing = "deferred"
			}
		}

		// Cross-check security classification from Inst lines.
		instSecurity := parseInstSecurity(simResult.Stdout)
		for i := range result.Updates {
			if result.Updates[i].Type != checker.UpdateTypeSecurity && instSecurity[result.Updates[i].Name] {
				result.Updates[i].Type = checker.UpdateTypeSecurity
				result.Updates[i].Priority = checker.PriorityHigh
			}
		}
	}

	// Detect kept-back packages via upgrade dry-run.
	// apt upgrade --dry-run (not full-upgrade) reports packages that require
	// new dependencies or removals as "kept back".
	slog.Info("detecting kept-back packages via upgrade dry-run")
	upgradeResult, err := executil.Run("apt", "upgrade", "--dry-run")
	if err != nil {
		slog.Warn("apt upgrade --dry-run failed, skipping kept-back detection", "error", err)
	} else {
		keptBack := parseKeptBackPackages(upgradeResult.Stdout)
		for i := range result.Updates {
			if keptBack[result.Updates[i].Name] {
				result.Updates[i].Phasing = "held"
			}
		}
	}

	// Filter out phased updates if configured.
	if a.hidePhased {
		var phasedCount int
		filtered := result.Updates[:0]
		for _, u := range result.Updates {
			if u.Phasing == "" || u.Phasing == "held" {
				filtered = append(filtered, u)
			} else {
				phasedCount++
			}
		}
		result.Updates = filtered
		if phasedCount > 0 {
			result.Notes = append(result.Notes,
				fmt.Sprintf("%d phased update(s) hidden", phasedCount))
		}
	}

	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
