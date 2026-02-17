package macos

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

var (
	labelRe       = regexp.MustCompile(`^\*\s+Label:\s+(.+)$`)
	versionRe     = regexp.MustCompile(`Version:\s+([^,]+)`)
	recommendedRe = regexp.MustCompile(`Recommended:\s+(YES|NO)`)
)

func parseSoftwareUpdate(output string, securityOnly bool) []checker.Update {
	var updates []checker.Update
	lines := strings.Split(output, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		matches := labelRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		label := matches[1]

		// Find the detail line (next non-empty line, typically tab-indented).
		var detailLine string
		for j := i + 1; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed != "" {
				detailLine = trimmed
				i = j
				break
			}
		}

		version := ""
		if vm := versionRe.FindStringSubmatch(detailLine); vm != nil {
			version = strings.TrimSpace(vm[1])
		}

		recommended := false
		if rm := recommendedRe.FindStringSubmatch(detailLine); rm != nil {
			recommended = rm[1] == "YES"
		}

		isSecurity := strings.Contains(label, "Security Update") ||
			strings.Contains(label, "Security Response")

		if securityOnly && !isSecurity {
			continue
		}

		updateType := checker.UpdateTypeRegular
		priority := checker.PriorityNormal
		if isSecurity {
			updateType = checker.UpdateTypeSecurity
			priority = checker.PriorityCritical
		} else if recommended {
			priority = checker.PriorityHigh
		}

		updates = append(updates, checker.Update{
			Name:       label,
			NewVersion: version,
			Type:       updateType,
			Priority:   priority,
		})
	}

	return updates
}
