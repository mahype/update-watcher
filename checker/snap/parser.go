package snap

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// snap refresh --list output format:
// Name           Version          Rev    Size    Publisher       Notes
// firefox        125.0.3-2        4336   283MB   mozilla**       -
var refreshLineRe = regexp.MustCompile(
	`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.+?)\s+(\S+)\s*$`,
)

// parseRefreshList parses the output of "snap refresh --list" into Updates.
func parseRefreshList(output string) []checker.Update {
	var updates []checker.Update
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip header line
		if i == 0 && strings.HasPrefix(line, "Name") {
			continue
		}

		matches := refreshLineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		updates = append(updates, checker.Update{
			Name:       matches[1],
			NewVersion: matches[2],
			Type:       checker.UpdateTypeRegular,
			Priority:   checker.PriorityNormal,
		})
	}

	return updates
}
