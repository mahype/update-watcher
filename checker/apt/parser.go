package apt

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// upgradable line format:
// libssl3/jammy-security 3.0.13-0ubuntu3.4 amd64 [upgradable from: 3.0.13-0ubuntu3.1]
// curl/jammy-updates 8.5.0-6 amd64 [upgradable from: 8.5.0-2] [phased 50%]
var upgradableRe = regexp.MustCompile(
	`^(\S+)/(\S+)\s+(\S+)\s+\S+\s+\[upgradable from:\s+(\S+)\](?:\s+\[phased\s+(\d+%)\])?`,
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

		var phasing string
		if len(matches) > 5 && matches[5] != "" {
			phasing = matches[5]
		}

		updates = append(updates, checker.Update{
			Name:           pkgName,
			CurrentVersion: currentVersion,
			NewVersion:     newVersion,
			Type:           updateType,
			Priority:       priority,
			Phasing:        phasing,
		})
	}

	return updates
}

// deferredRe matches the "deferred due to phasing" line from apt-get -s upgrade output.
var deferredRe = regexp.MustCompile(
	`The following upgrades have been deferred due to phasing:\s*\n((?:\s+\S.*\n?)*)`,
)

// parseDeferredPackages extracts package names from "apt-get -s upgrade" output
// that are deferred due to phasing. Returns a set of package names.
func parseDeferredPackages(output string) map[string]bool {
	deferred := make(map[string]bool)

	match := deferredRe.FindStringSubmatch(output)
	if match == nil {
		return deferred
	}

	for _, name := range strings.Fields(match[1]) {
		deferred[name] = true
	}

	return deferred
}
