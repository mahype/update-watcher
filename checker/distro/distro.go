package distro

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func init() {
	checker.Register("distro", NewFromConfig)
}

// distroBackend is the internal interface for distribution-specific upgrade checks.
type distroBackend interface {
	// CheckUpgrade checks if a newer release is available.
	// Returns the new version string and true if available, or empty string and false.
	CheckUpgrade(current OSRelease) (newVersion string, available bool, err error)

	// UpgradeCommand returns the command a user would run to perform the upgrade.
	UpgradeCommand() string
}

// DistroChecker checks for available distribution release upgrades.
type DistroChecker struct {
	ltsOnly     bool
	osReleaseFile string
}

// NewFromConfig creates a DistroChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &DistroChecker{
		ltsOnly:       cfg.GetBool("lts_only", true),
		osReleaseFile: cfg.GetString("os_release_file", "/etc/os-release"),
	}, nil
}

func (c *DistroChecker) Name() string { return "distro" }

func (c *DistroChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: c.Name(),
		CheckedAt:   time.Now(),
	}

	// Parse os-release
	osRel, err := ParseOSRelease(c.osReleaseFile)
	if err != nil {
		return result, fmt.Errorf("failed to read os-release: %w", err)
	}

	slog.Info("detected distribution", "id", osRel.ID, "version", osRel.VersionID, "name", osRel.PrettyName)

	// Select backend
	backend := c.selectBackend(osRel)
	if backend == nil {
		result.Summary = fmt.Sprintf("distro release check not supported for %s", osRel.ID)
		slog.Info("no distro backend available", "id", osRel.ID)
		return result, nil
	}

	// Check for upgrade
	newVersion, available, err := backend.CheckUpgrade(osRel)
	if err != nil {
		return result, fmt.Errorf("upgrade check failed: %w", err)
	}

	if available {
		displayName := osRel.Name
		if displayName == "" {
			displayName = osRel.ID
		}
		result.Updates = []checker.Update{
			{
				Name:           displayName,
				CurrentVersion: osRel.VersionID,
				NewVersion:     newVersion,
				Type:           checker.UpdateTypeDistro,
				Priority:       checker.PriorityHigh,
				Source:         backend.UpgradeCommand(),
			},
		}
		result.Summary = fmt.Sprintf("release upgrade available: %s \u2192 %s", osRel.VersionID, newVersion)
	} else {
		result.Summary = fmt.Sprintf("running latest release (%s)", osRel.VersionID)
	}

	return result, nil
}

// selectBackend returns the appropriate backend for the detected distribution.
// Returns nil if the distribution is not supported.
func (c *DistroChecker) selectBackend(osRel OSRelease) distroBackend {
	switch osRel.ID {
	case "ubuntu":
		return &ubuntuBackend{ltsOnly: c.ltsOnly}
	case "debian":
		return &debianBackend{}
	case "fedora":
		return &fedoraBackend{}
	}

	// Fallback: check ID_LIKE for derivative distributions
	for _, like := range strings.Fields(osRel.IDLike) {
		switch like {
		case "ubuntu":
			return &ubuntuBackend{ltsOnly: c.ltsOnly}
		case "debian":
			return &debianBackend{}
		case "fedora":
			return &fedoraBackend{}
		}
	}

	return nil
}
