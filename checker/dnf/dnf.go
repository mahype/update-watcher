package dnf

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
	checker.Register("dnf", NewFromConfig)
}

// DnfChecker checks for available DNF package updates (Fedora/RHEL/Rocky/AlmaLinux).
type DnfChecker struct {
	useSudo      bool
	securityOnly bool
}

// NewFromConfig creates a DnfChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &DnfChecker{
		useSudo:      cfg.GetBool("use_sudo", true),
		securityOnly: cfg.GetBool("security_only", false),
	}, nil
}

func (d *DnfChecker) Name() string { return "dnf" }

func (d *DnfChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: d.Name(),
		CheckedAt:   time.Now(),
	}

	// Get security updates first (if needed for classification)
	var securityPkgs map[string]bool
	if !d.securityOnly {
		slog.Info("fetching security update info")
		var secResult *executil.Result
		var err error
		secResult, err = executil.RunMaybeSudo(d.useSudo, "dnf", "updateinfo", "list", "--security", "-q")
		if err != nil {
			slog.Warn("dnf updateinfo failed, security classification unavailable", "error", err)
		} else {
			securityPkgs = parseSecurityInfo(secResult.Stdout)
		}
	}

	// Check for updates — dnf check-update returns exit code 100 when updates are available
	slog.Info("checking for available updates")
	var checkResult *executil.Result
	var err error

	checkResult, err = executil.RunMaybeSudo(d.useSudo, "dnf", "check-update", "-q")

	// Exit code 100 means updates are available (not an error)
	if err != nil && checkResult != nil && checkResult.ExitCode != 100 {
		return result, fmt.Errorf("dnf check-update failed: %w", err)
	}

	if d.securityOnly {
		// For security-only mode, use dnf updateinfo directly
		var secResult *executil.Result
		secResult, err = executil.RunMaybeSudo(d.useSudo, "dnf", "updateinfo", "list", "--security", "-q")
		if err != nil {
			return result, fmt.Errorf("dnf updateinfo failed: %w", err)
		}
		result.Updates = parseSecurityUpdates(secResult.Stdout)
	} else {
		result.Updates = parseCheckUpdate(checkResult.Stdout, securityPkgs)
	}

	result.Summary = checker.BuildSummary(result.Updates, "packages")

	return result, nil
}
