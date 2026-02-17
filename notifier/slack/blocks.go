package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
)

const maxUpdatesPerSection = 10

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
	totalUpdates := 0
	totalSecurity := 0
	checkerCount := len(results)
	for _, r := range results {
		totalUpdates += len(r.Updates)
		for _, u := range r.Updates {
			if u.Type == checker.UpdateTypeSecurity {
				totalSecurity++
			}
		}
	}

	contextText := fmt.Sprintf("Checked at %s  |  %d checkers  |  %d updates found",
		time.Now().UTC().Format("2006-01-02 15:04 UTC"), checkerCount, totalUpdates)
	blocks = append(blocks, contextBlock(contextText))

	// Per-checker sections
	for _, r := range results {
		blocks = append(blocks, dividerBlock())

		icon := checkerEmoji(r.CheckerName, useEmoji)
		sectionTitle := fmt.Sprintf("*%s %s* — %s", icon, checkerDisplayName(r.CheckerName), r.Summary)

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
	if totalSecurity > 0 {
		blocks = append(blocks, dividerBlock())
		blocks = append(blocks, contextBlock(fmt.Sprintf("<!channel> Security updates require attention  |  _update-watcher_")))
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
	shown := 0
	for _, u := range r.Updates {
		if shown >= maxUpdatesPerSection {
			lines = append(lines, fmt.Sprintf("_...and %d more_", len(r.Updates)-maxUpdatesPerSection))
			break
		}
		indicator := priorityIndicator(u, useEmoji)
		line := fmt.Sprintf("%s `%s` %s \u2192 %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		if u.Type == checker.UpdateTypeSecurity {
			line += " _(security)_"
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", u.Source)
		}
		lines = append(lines, line)
		shown++
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
		shown := 0
		for _, u := range siteUpdates {
			if shown >= maxUpdatesPerSection {
				lines = append(lines, fmt.Sprintf("_...and %d more_", len(siteUpdates)-maxUpdatesPerSection))
				break
			}
			indicator := priorityIndicator(u, useEmoji)
			typeName := strings.Title(u.Type)
			line := fmt.Sprintf("%s %s: `%s` %s \u2192 %s", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			if u.Type == checker.UpdateTypeSecurity {
				line += " _(security)_"
			}
			lines = append(lines, line)
			shown++
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func checkerEmoji(name string, useEmoji bool) string {
	if !useEmoji {
		return ""
	}
	switch name {
	case "apt":
		return "\U0001f427" // 🐧
	case "docker":
		return "\U0001f433" // 🐳
	case "wordpress":
		return "\U0001f4dd" // 📝
	default:
		return "\U0001f504" // 🔄
	}
}

func checkerDisplayName(name string) string {
	switch name {
	case "apt":
		return "APT Updates"
	case "docker":
		return "Docker Updates"
	case "wordpress":
		return "WordPress Updates"
	default:
		return name + " Updates"
	}
}

func priorityIndicator(u checker.Update, useEmoji bool) string {
	if !useEmoji {
		if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
			return "[!]"
		}
		return "[-]"
	}
	if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
		return "\U0001f534" // 🔴
	}
	return "\u26aa" // ⚪
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
