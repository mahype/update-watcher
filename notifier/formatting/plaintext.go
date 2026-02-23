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

		if cmd := UpdateCommandForResult(r.CheckerName, r.Updates); cmd != "" && len(r.Updates) > 0 {
			parts = append(parts, fmt.Sprintf("  -> Update: %s", cmd))
		}

		if count, cmd := PhasingNote(r.CheckerName, r.Updates); count > 0 {
			parts = append(parts, fmt.Sprintf("  -> %d phased update(s) cannot be installed via regular upgrade. Use: %s", count, cmd))
		}

		if count, cmd := KeptBackNote(r.CheckerName, r.Updates); count > 0 {
			parts = append(parts, fmt.Sprintf("  -> %d package(s) held back \u2014 need new dependencies or removals. Use: %s", count, cmd))
		}

		for _, note := range r.Notes {
			parts = append(parts, fmt.Sprintf("  -> %s", note))
		}

		parts = append(parts, "")
	}

	// Security footer
	if summary.SecurityCount > 0 {
		parts = append(parts, fmt.Sprintf("! Security updates require attention (%d security updates)", summary.SecurityCount))
	}

	return strings.Join(parts, "\n")
}
