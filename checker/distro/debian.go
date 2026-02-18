package distro

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mahype/update-watcher/internal/executil"
)

// debianBackend checks for Debian release upgrades.
type debianBackend struct{}

var debianVersionRe = regexp.MustCompile(`(?m)^Version:\s*(\d+)`)

func (d *debianBackend) CheckUpgrade(current OSRelease) (string, bool, error) {
	currentMajor, err := parseMajorVersion(current.VersionID)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse current Debian version %q: %w", current.VersionID, err)
	}

	// Strategy 1: use debian-distro-info if available
	if _, lookErr := exec.LookPath("debian-distro-info"); lookErr == nil {
		return d.checkViaDistroInfo(currentMajor)
	}

	// Strategy 2: HTTP fallback
	return d.checkViaHTTP(currentMajor)
}

func (d *debianBackend) UpgradeCommand() string {
	return "See https://www.debian.org/releases/"
}

func (d *debianBackend) checkViaDistroInfo(currentMajor int) (string, bool, error) {
	result, err := executil.RunWithTimeout(10*time.Second, "debian-distro-info", "-r", "--stable")
	if err != nil {
		return "", false, fmt.Errorf("debian-distro-info failed: %w", err)
	}

	stableVersion := strings.TrimSpace(result.Stdout)
	stableMajor, err := parseMajorVersion(stableVersion)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse stable version %q: %w", stableVersion, err)
	}

	if stableMajor > currentMajor {
		return stableVersion, true, nil
	}
	return "", false, nil
}

func (d *debianBackend) checkViaHTTP(currentMajor int) (string, bool, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://deb.debian.org/debian/dists/stable/Release")
	if err != nil {
		return "", false, fmt.Errorf("fetching Debian release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("Debian release info returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", false, fmt.Errorf("reading Debian release info: %w", err)
	}

	stableVersion := parseDebianReleaseVersion(string(body))
	if stableVersion == "" {
		return "", false, fmt.Errorf("could not parse Version from Debian Release file")
	}

	stableMajor, err := parseMajorVersion(stableVersion)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse stable version %q: %w", stableVersion, err)
	}

	if stableMajor > currentMajor {
		return stableVersion, true, nil
	}
	return "", false, nil
}

// parseDebianReleaseVersion extracts the Version field from a Debian Release file.
func parseDebianReleaseVersion(content string) string {
	m := debianVersionRe.FindStringSubmatch(content)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// parseMajorVersion extracts the leading integer from a version string.
// For example "12.5" -> 12, "22.04" -> 22.
func parseMajorVersion(version string) (int, error) {
	parts := strings.SplitN(version, ".", 2)
	return strconv.Atoi(parts[0])
}
