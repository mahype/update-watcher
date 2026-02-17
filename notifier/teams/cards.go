package teams

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/notifier/formatting"
)

// BuildAdaptiveCard creates a Teams Adaptive Card v1.5 from check results.
func BuildAdaptiveCard(hostname string, results []*checker.CheckResult) map[string]interface{} {
	summary := formatting.SummarizeResults(results)

	var bodyElements []map[string]interface{}

	// Header
	bodyElements = append(bodyElements, map[string]interface{}{
		"type":   "TextBlock",
		"size":   "Large",
		"weight": "Bolder",
		"text":   fmt.Sprintf("\U0001f504 Update Report: %s", hostname),
	})

	// Context
	bodyElements = append(bodyElements, map[string]interface{}{
		"type":     "TextBlock",
		"text":     fmt.Sprintf("Checked at %s", time.Now().UTC().Format("2006-01-02 15:04 UTC")),
		"isSubtle": true,
		"spacing":  "None",
	})

	// Summary facts
	facts := []map[string]interface{}{
		{"title": "Checkers", "value": fmt.Sprintf("%d", summary.CheckerCount)},
		{"title": "Updates", "value": fmt.Sprintf("%d", summary.TotalUpdates)},
	}
	if summary.SecurityCount > 0 {
		facts = append(facts, map[string]interface{}{
			"title": "Security",
			"value": fmt.Sprintf("\u26a0\ufe0f %d", summary.SecurityCount),
		})
	}
	bodyElements = append(bodyElements, map[string]interface{}{
		"type":  "FactSet",
		"facts": facts,
	})

	// Per-checker containers
	for _, r := range results {
		icon := formatting.CheckerEmoji(r.CheckerName, true)
		items := []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   fmt.Sprintf("**%s %s** — %s", icon, formatting.CheckerDisplayName(r.CheckerName), r.Summary),
				"weight": "Bolder",
				"wrap":   true,
			},
		}

		if r.Error != "" {
			items = append(items, map[string]interface{}{
				"type":  "TextBlock",
				"text":  fmt.Sprintf("\u26a0\ufe0f %s", r.Error),
				"color": "Attention",
				"wrap":  true,
			})
		}

		updates := formatUpdatesTeams(r)
		if updates != "" {
			items = append(items, map[string]interface{}{
				"type": "TextBlock",
				"text": updates,
				"wrap": true,
			})
		}

		bodyElements = append(bodyElements, map[string]interface{}{
			"type":      "Container",
			"separator": true,
			"items":     items,
		})
	}

	// Security footer
	if summary.SecurityCount > 0 {
		bodyElements = append(bodyElements, map[string]interface{}{
			"type":      "TextBlock",
			"text":      fmt.Sprintf("\u26a0\ufe0f **Security updates require attention** (%d security updates)", summary.SecurityCount),
			"color":     "Attention",
			"weight":    "Bolder",
			"separator": true,
			"wrap":      true,
		})
	}

	return map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.5",
		"body":    bodyElements,
	}
}

func formatUpdatesTeams(r *checker.CheckResult) string {
	if len(r.Updates) == 0 {
		return ""
	}

	if r.CheckerName == "wordpress" {
		return formatWordPressUpdatesTeams(r.Updates)
	}

	var lines []string
	shown := 0
	for _, u := range r.Updates {
		if shown >= formatting.MaxUpdatesPerSection {
			lines = append(lines, fmt.Sprintf("*...and %d more*", len(r.Updates)-formatting.MaxUpdatesPerSection))
			break
		}
		indicator := formatting.PriorityIndicator(u, true)
		line := fmt.Sprintf("%s `%s` %s \u2192 %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		if u.Type == checker.UpdateTypeSecurity {
			line += " *(security)*"
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", u.Source)
		}
		lines = append(lines, line)
		shown++
	}

	return strings.Join(lines, "\n\n")
}

func formatWordPressUpdatesTeams(updates []checker.Update) string {
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
		lines := []string{fmt.Sprintf("**%s**", source)}
		shown := 0
		for _, u := range siteUpdates {
			if shown >= formatting.MaxUpdatesPerSection {
				lines = append(lines, fmt.Sprintf("*...and %d more*", len(siteUpdates)-formatting.MaxUpdatesPerSection))
				break
			}
			indicator := formatting.PriorityIndicator(u, true)
			typeName := strings.ToUpper(u.Type[:1]) + u.Type[1:]
			line := fmt.Sprintf("%s %s: `%s` %s \u2192 %s", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			if u.Type == checker.UpdateTypeSecurity {
				line += " *(security)*"
			}
			lines = append(lines, line)
			shown++
		}
		sections = append(sections, strings.Join(lines, "\n\n"))
	}

	return strings.Join(sections, "\n\n")
}
