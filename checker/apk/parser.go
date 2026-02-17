package apk

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// apk version -l '<' output format:
// Installed:                          Available:
// busybox-1.36.1-r15                < 1.36.1-r19
// curl-8.5.0-r0                     < 8.6.0-r0
// openssl-3.1.4-r2                  < 3.1.4-r5
//
// Or without header (depends on version):
// busybox-1.36.1-r15 < 1.36.1-r19
var versionRe = regexp.MustCompile(
	`^(\S+?)-(\d\S*?)\s+<\s+(\S+)\s*$`,
)

// parseVersionOutput parses the output of "apk version -l '<'" into Updates.
func parseVersionOutput(output string) []checker.Update {
	var updates []checker.Update

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip header line
		if strings.HasPrefix(line, "Installed:") {
			continue
		}

		matches := versionRe.FindStringSubmatch(line)
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
