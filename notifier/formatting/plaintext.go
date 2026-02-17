package formatting

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
)

// BuildPlainTextMessage creates a plain text update report.
func BuildPlainTextMessage(hostname string, results []*checker.CheckResult) string {
	summary := SummarizeResults(results)

	var parts []string

	// Header
	parts = append(parts, fmt.Sprintf("Update Report: %s", hostname))
	parts = append(parts, fmt.Sprintf("Checked at %s | %d checkers | %d updates found",
		time.Now().UTC().Format("2006-01-02 15:04 UTC"), summary.CheckerCount, summary.TotalUpdates))
	parts = append(parts, "")

	// Per-checker sections
	for _, r := range results {
		sectionTitle := fmt.Sprintf("%s — %s", CheckerDisplayName(r.CheckerName), r.Summary)
		parts = append(parts, sectionTitle)

		if r.Error != "" {
			parts = append(parts, fmt.Sprintf("  WARNING: %s", r.Error))
		}

		updates := FormatUpdatesPlainText(r)
		if updates != "" {
			parts = append(parts, updates)
		}

		if cmd := UpdateCommand(r.CheckerName); cmd != "" && len(r.Updates) > 0 {
			parts = append(parts, fmt.Sprintf("  -> Update: %s", cmd))
		}

		parts = append(parts, "")
	}

	// Security footer
	if summary.SecurityCount > 0 {
		parts = append(parts, fmt.Sprintf("! Security updates require attention (%d security updates)", summary.SecurityCount))
	}

	return strings.Join(parts, "\n")
}
