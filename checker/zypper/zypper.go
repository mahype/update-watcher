package zypper

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
	checker.Register("zypper", NewFromConfig)
}

// ZypperChecker checks for available Zypper package updates (openSUSE/SLES).
type ZypperChecker struct {
	useSudo      bool
	securityOnly bool
}

// NewFromConfig creates a ZypperChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &ZypperChecker{
		useSudo:      cfg.GetBool("use_sudo", true),
		securityOnly: cfg.GetBool("security_only", false),
	}, nil
}

func (z *ZypperChecker) Name() string { return "zypper" }

func (z *ZypperChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: z.Name(),
		CheckedAt:   time.Now(),
	}

	// Refresh repository data
	slog.Info("refreshing zypper repositories")
	var refreshResult *executil.Result
	var err error

	refreshResult, err = executil.RunMaybeSudo(z.useSudo, "zypper", "--non-interactive", "refresh")
	if err != nil {
		slog.Warn("zypper refresh failed, continuing with possibly stale data", "error", err, "stderr", refreshResult.Stderr)
		result.Error = fmt.Sprintf("zypper refresh failed: %s", err)
	}

	// Get security patches info for classification
	var securityPkgs map[string]bool
	slog.Info("checking for security patches")
	var patchResult *executil.Result
	patchResult, err = executil.RunMaybeSudo(z.useSudo, "zypper", "--non-interactive", "list-patches", "--category", "security")
	if err != nil {
		slog.Warn("zypper list-patches failed, security classification unavailable", "error", err)
	} else {
		securityPkgs = parseSecurityPatches(patchResult.Stdout)
	}

	// List available updates
	slog.Info("checking for available updates")
	var listResult *executil.Result
	listResult, err = executil.RunMaybeSudo(z.useSudo, "zypper", "--non-interactive", "list-updates")
	if err != nil {
		return result, fmt.Errorf("zypper list-updates failed: %w", err)
	}

	allUpdates := parseListUpdates(listResult.Stdout, securityPkgs)

	if z.securityOnly {
		for _, u := range allUpdates {
			if u.Type == checker.UpdateTypeSecurity {
				result.Updates = append(result.Updates, u)
			}
		}
	} else {
		result.Updates = allUpdates
	}

	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
