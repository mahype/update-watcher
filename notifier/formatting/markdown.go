package formatting

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
)

// MessageOptions controls message formatting behavior.
type MessageOptions struct {
	UseEmoji bool
}

// DefaultOptions returns sensible default formatting options.
func DefaultOptions() MessageOptions {
	return MessageOptions{UseEmoji: true}
}

// BuildMarkdownMessage creates a markdown-formatted update report.
// Returns title and body as separate strings.
func BuildMarkdownMessage(hostname string, results []*checker.CheckResult, opts MessageOptions) (title string, body string) {
	summary := SummarizeResults(results)

	// Title
	title = "Update Report: " + hostname
	if opts.UseEmoji {
		title = "\U0001f504 " + title // 🔄
	}

	var parts []string

	// Context line
	contextLine := fmt.Sprintf("Checked at %s | %d checkers | %d updates found",
		time.Now().UTC().Format("2006-01-02 15:04 UTC"), summary.CheckerCount, summary.TotalUpdates)
	parts = append(parts, contextLine)

	// Per-checker sections
	for _, r := range results {
		icon := CheckerEmoji(r.CheckerName, opts.UseEmoji)
		sectionTitle := fmt.Sprintf("---\n### %s %s — %s", icon, CheckerDisplayName(r.CheckerName), r.Summary)

		if r.Error != "" {
			sectionTitle += fmt.Sprintf("\n\u26a0\ufe0f %s", r.Error)
		}

		updates := FormatUpdatesMarkdown(r, opts.UseEmoji)
		if updates != "" {
			sectionTitle += "\n\n" + updates
		}

		if cmd := UpdateCommand(r.CheckerName); cmd != "" && len(r.Updates) > 0 {
			sectionTitle += fmt.Sprintf("\n\n> \U0001f4a1 Update: `%s`", cmd)
		}

		if count, cmd := PhasingNote(r.CheckerName, r.Updates); count > 0 {
			sectionTitle += fmt.Sprintf("\n> \u23f3 %d phased update(s) cannot be installed via regular upgrade. Use:\n> `%s`", count, cmd)
		}

		for _, note := range r.Notes {
			sectionTitle += fmt.Sprintf("\n> \u23f3 %s", note)
		}

		parts = append(parts, sectionTitle)
	}

	// Security footer
	if summary.SecurityCount > 0 {
		footer := fmt.Sprintf("---\n\u26a0\ufe0f **Security updates require attention** (%d security updates)", summary.SecurityCount)
		parts = append(parts, footer)
	}

	body = strings.Join(parts, "\n\n")
	return title, body
}
