package distro

import (
	"regexp"
	"time"

	"github.com/mahype/update-watcher/internal/executil"
)

// ubuntuBackend checks for Ubuntu release upgrades using do-release-upgrade.
type ubuntuBackend struct {
	ltsOnly bool
}

var ubuntuUpgradeRe = regexp.MustCompile(`New release '([^']+)' available`)

func (u *ubuntuBackend) CheckUpgrade(current OSRelease) (string, bool, error) {
	// do-release-upgrade -c checks for available release upgrades.
	// Exit code 0 means an upgrade is available; non-zero means none.
	// On LTS systems the default behavior is to only show LTS upgrades,
	// which matches ltsOnly=true. The tool respects /etc/update-manager/release-upgrades.
	result, err := executil.RunWithTimeout(30*time.Second, "do-release-upgrade", "-c")
	if err != nil {
		// Non-zero exit code typically means no upgrade available.
		if result != nil && result.ExitCode != 0 {
			return "", false, nil
		}
		return "", false, err
	}

	newVersion := parseUbuntuUpgradeOutput(result.Stdout + "\n" + result.Stderr)
	if newVersion == "" {
		return "", false, nil
	}
	return newVersion, true, nil
}

func (u *ubuntuBackend) UpgradeCommand() string {
	return "sudo do-release-upgrade"
}

// parseUbuntuUpgradeOutput extracts the new version from do-release-upgrade -c output.
// Example: "New release '24.04 LTS' available."
func parseUbuntuUpgradeOutput(output string) string {
	m := ubuntuUpgradeRe.FindStringSubmatch(output)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
