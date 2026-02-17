package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/notifier/formatting"
)

// Block is a Slack Block Kit block.
type Block map[string]interface{}

// BuildMessage creates a Slack Block Kit message from check results.
func BuildMessage(hostname string, results []*checker.CheckResult, useEmoji bool) []Block {
	var blocks []Block

	// Header
	title := "Update Report: " + hostname
	if useEmoji {
		title = "\U0001f504 " + title // 🔄
	}
	blocks = append(blocks, headerBlock(title))

	// Context: timestamp and summary
	summary := formatting.SummarizeResults(results)

	contextText := fmt.Sprintf("Checked at %s  |  %d checkers  |  %d updates found",
		time.Now().UTC().Format("2006-01-02 15:04 UTC"), summary.CheckerCount, summary.TotalUpdates)
	blocks = append(blocks, contextBlock(contextText))

	// Per-checker sections
	for _, r := range results {
		blocks = append(blocks, dividerBlock())

		icon := formatting.CheckerEmoji(r.CheckerName, useEmoji)
		sectionTitle := fmt.Sprintf("*%s %s* — %s", icon, formatting.CheckerDisplayName(r.CheckerName), r.Summary)

		if r.Error != "" {
			sectionTitle += fmt.Sprintf("\n:warning: %s", r.Error)
		}

		body := formatUpdates(r, useEmoji)
		if body != "" {
			sectionTitle += "\n\n" + body
		}

		blocks = append(blocks, sectionBlock(sectionTitle))
	}

	// Footer with mentions if security updates present
	if summary.SecurityCount > 0 {
		blocks = append(blocks, dividerBlock())
		blocks = append(blocks, contextBlock("<!channel> Security updates require attention  |  _update-watcher_"))
	}

	return blocks
}

func formatUpdates(r *checker.CheckResult, useEmoji bool) string {
	if len(r.Updates) == 0 {
		return ""
	}

	// Group WordPress updates by source (site name)
	if r.CheckerName == "wordpress" {
		return formatWordPressUpdates(r.Updates, useEmoji)
	}

	var lines []string
	for _, u := range r.Updates {
		indicator := formatting.PriorityIndicator(u, useEmoji)
		var line string
		if u.Type == checker.UpdateTypeSecurity {
			line = fmt.Sprintf("%s *`%s`* %s \u2192 %s \u26a0\ufe0f *SECURITY*", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		} else {
			line = fmt.Sprintf("%s `%s` %s \u2192 %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", u.Source)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWordPressUpdates(updates []checker.Update, useEmoji bool) string {
	// Group by source (site name)
	grouped := make(map[string][]checker.Update)
	var order []string
	for _, u := range updates {
		if _, exists := grouped[u.Source]; !exists {
			order = append(order, u.Source)
		}
		grouped[u.Source] = append(grouped[u.Source], u)
	}

	var sections []string
	for _, source := range order {
		siteUpdates := grouped[source]
		lines := []string{fmt.Sprintf("*%s*", source)}
		for _, u := range siteUpdates {
			indicator := formatting.PriorityIndicator(u, useEmoji)
			typeName := strings.ToUpper(u.Type[:1]) + u.Type[1:]
			var line string
			if u.Type == checker.UpdateTypeSecurity {
				line = fmt.Sprintf("%s %s: *`%s`* %s \u2192 %s \u26a0\ufe0f *SECURITY*", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			} else {
				line = fmt.Sprintf("%s %s: `%s` %s \u2192 %s", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			}
			lines = append(lines, line)
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func headerBlock(text string) Block {
	return Block{
		"type": "header",
		"text": Block{
			"type":  "plain_text",
			"text":  text,
			"emoji": true,
		},
	}
}

func contextBlock(text string) Block {
	return Block{
		"type": "context",
		"elements": []Block{
			{"type": "mrkdwn", "text": text},
		},
	}
}

func dividerBlock() Block {
	return Block{"type": "divider"}
}

func sectionBlock(text string) Block {
	return Block{
		"type": "section",
		"text": Block{
			"type": "mrkdwn",
			"text": text,
		},
	}
}
