package pacman

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// pacman -Qu output format:
// package-name current-version -> new-version
// Example:
// linux 6.6.7.arch1-1 -> 6.6.8.arch1-1
var upgradableRe = regexp.MustCompile(
	`^(\S+)\s+(\S+)\s+->\s+(\S+)\s*$`,
)

// parseUpgradable parses the output of "pacman -Qu" into Updates.
func parseUpgradable(output string) []checker.Update {
	var updates []checker.Update

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip ignored packages
		if strings.Contains(line, "[ignored]") {
			continue
		}

		matches := upgradableRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		updates = append(updates, checker.Update{
			Name:           matches[1],
			CurrentVersion: matches[2],
			NewVersion:     matches[3],
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
		})
	}

	return updates
}
