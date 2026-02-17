package flatpak

import (
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// parseRemoteUpdates parses the tab-separated output of
// "flatpak remote-ls --updates --app --columns=name,application,version".
func parseRemoteUpdates(output string) []checker.Update {
	var updates []checker.Update

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		appID := strings.TrimSpace(parts[1])
		version := ""
		if len(parts) >= 3 {
			version = strings.TrimSpace(parts[2])
		}

		// Use application ID as the package name (unique identifier)
		updates = append(updates, checker.Update{
			Name:       appID,
			NewVersion: version,
			Type:       checker.UpdateTypeRegular,
			Priority:   checker.PriorityNormal,
			Source:     name, // human-readable name
		})
	}

	return updates
}
