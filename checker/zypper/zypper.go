package zypper

import (
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

func (z *ZypperChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: z.Name(),
		CheckedAt:   time.Now(),
	}

	// Refresh repository data
	slog.Info("refreshing zypper repositories")
	var refreshResult *executil.Result
	var err error

	if z.useSudo {
		refreshResult, err = executil.RunAsSudo("zypper", "--non-interactive", "refresh")
	} else {
		refreshResult, err = executil.Run("zypper", "--non-interactive", "refresh")
	}
	if err != nil {
		slog.Warn("zypper refresh failed, continuing with possibly stale data", "error", err, "stderr", refreshResult.Stderr)
		result.Error = fmt.Sprintf("zypper refresh failed: %s", err)
	}

	// Get security patches info for classification
	var securityPkgs map[string]bool
	slog.Info("checking for security patches")
	var patchResult *executil.Result
	if z.useSudo {
		patchResult, err = executil.RunAsSudo("zypper", "--non-interactive", "list-patches", "--category", "security")
	} else {
		patchResult, err = executil.Run("zypper", "--non-interactive", "list-patches", "--category", "security")
	}
	if err != nil {
		slog.Warn("zypper list-patches failed, security classification unavailable", "error", err)
	} else {
		securityPkgs = parseSecurityPatches(patchResult.Stdout)
	}

	// List available updates
	slog.Info("checking for available updates")
	var listResult *executil.Result
	if z.useSudo {
		listResult, err = executil.RunAsSudo("zypper", "--non-interactive", "list-updates")
	} else {
		listResult, err = executil.Run("zypper", "--non-interactive", "list-updates")
	}
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
