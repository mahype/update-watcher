package distro

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// fedoraBackend checks for Fedora release upgrades via the releases API.
type fedoraBackend struct{}

// fedoraRelease represents a single entry from the Fedora releases JSON.
type fedoraRelease struct {
	Version string `json:"version"`
	Stable  string `json:"stable"`
}

func (f *fedoraBackend) CheckUpgrade(current OSRelease) (string, bool, error) {
	currentVersion, err := strconv.Atoi(current.VersionID)
	if err != nil {
		return "", false, fmt.Errorf("cannot parse current Fedora version %q: %w", current.VersionID, err)
	}

	latestStable, err := fetchLatestFedoraRelease()
	if err != nil {
		return "", false, err
	}

	if latestStable > currentVersion {
		return strconv.Itoa(latestStable), true, nil
	}
	return "", false, nil
}

func (f *fedoraBackend) UpgradeCommand() string {
	return "sudo dnf system-upgrade"
}

// fetchLatestFedoraRelease queries the Fedora releases API and returns the
// highest stable version number.
func fetchLatestFedoraRelease() (int, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://fedoraproject.org/releases.json")
	if err != nil {
		return 0, fmt.Errorf("fetching Fedora releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Fedora releases API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return 0, fmt.Errorf("reading Fedora releases: %w", err)
	}

	return parseLatestFedoraRelease(body)
}

// parseLatestFedoraRelease parses the Fedora releases JSON and returns the
// highest stable version number.
func parseLatestFedoraRelease(data []byte) (int, error) {
	var releases []fedoraRelease
	if err := json.Unmarshal(data, &releases); err != nil {
		return 0, fmt.Errorf("parsing Fedora releases JSON: %w", err)
	}

	maxVersion := 0
	for _, r := range releases {
		// Only consider releases that have reached stable.
		if r.Stable == "" {
			continue
		}
		v, err := strconv.Atoi(r.Version)
		if err != nil {
			continue
		}
		if v > maxVersion {
			maxVersion = v
		}
	}

	if maxVersion == 0 {
		return 0, fmt.Errorf("no stable Fedora releases found")
	}
	return maxVersion, nil
}
