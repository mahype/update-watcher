package apt

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// upgradable line format:
// libssl3/jammy-security 3.0.13-0ubuntu3.4 amd64 [upgradable from: 3.0.13-0ubuntu3.1]
var upgradableRe = regexp.MustCompile(
	`^(\S+)/(\S+)\s+(\S+)\s+\S+\s+\[upgradable from:\s+(\S+)\]`,
)

// parseUpgradable parses the output of "apt list --upgradable" into Updates.
func parseUpgradable(output string, securityOnly bool) []checker.Update {
	var updates []checker.Update

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Listing...") {
			continue
		}

		matches := upgradableRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		pkgName := matches[1]
		origin := matches[2]
		newVersion := matches[3]
		currentVersion := matches[4]

		isSecurity := strings.Contains(origin, "-security")

		if securityOnly && !isSecurity {
			continue
		}

		updateType := checker.UpdateTypeRegular
		priority := checker.PriorityNormal
		if isSecurity {
			updateType = checker.UpdateTypeSecurity
			priority = checker.PriorityHigh
		}

		updates = append(updates, checker.Update{
			Name:           pkgName,
			CurrentVersion: currentVersion,
			NewVersion:     newVersion,
			Type:           updateType,
			Priority:       priority,
		})
	}

	return updates
}
