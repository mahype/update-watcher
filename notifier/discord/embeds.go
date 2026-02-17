package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/notifier/formatting"
)

const (
	colorGreen  = 0x00FF00 // no updates
	colorYellow = 0xFFFF00 // regular updates
	colorRed    = 0xFF0000 // security updates
)

// Embed represents a Discord embed object.
type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
}

// Field represents a Discord embed field.
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Footer represents a Discord embed footer.
type Footer struct {
	Text string `json:"text"`
}

// BuildEmbeds creates Discord embeds from check results.
func BuildEmbeds(hostname string, results []*checker.CheckResult) []Embed {
	summary := formatting.SummarizeResults(results)

	// Determine color
	color := colorGreen
	if summary.SecurityCount > 0 {
		color = colorRed
	} else if summary.TotalUpdates > 0 {
		color = colorYellow
	}

	embed := Embed{
		Title: fmt.Sprintf("\U0001f504 Update Report: %s", hostname),
		Description: fmt.Sprintf("Checked at %s | %d checkers | %d updates found",
			time.Now().UTC().Format("2006-01-02 15:04 UTC"), summary.CheckerCount, summary.TotalUpdates),
		Color:     color,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Footer:    &Footer{Text: "update-watcher"},
	}

	// Add fields per checker
	for _, r := range results {
		icon := formatting.CheckerEmoji(r.CheckerName, true)
		name := fmt.Sprintf("%s %s", icon, formatting.CheckerDisplayName(r.CheckerName))

		value := r.Summary
		if r.Error != "" {
			value += fmt.Sprintf("\n\u26a0\ufe0f %s", r.Error)
		}

		updates := formatUpdatesDiscord(r)
		if updates != "" {
			value += "\n" + updates
		}

		// Discord field value limit is 1024 chars
		if len(value) > 1024 {
			value = value[:1020] + "..."
		}

		embed.Fields = append(embed.Fields, Field{
			Name:  name,
			Value: value,
		})
	}

	// Security footer in description
	if summary.SecurityCount > 0 {
		embed.Description += fmt.Sprintf("\n\n\u26a0\ufe0f **Security updates require attention** (%d security updates)", summary.SecurityCount)
	}

	return []Embed{embed}
}

func formatUpdatesDiscord(r *checker.CheckResult) string {
	if len(r.Updates) == 0 {
		return ""
	}

	if r.CheckerName == "wordpress" {
		return formatWordPressUpdatesDiscord(r.Updates)
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

	return strings.Join(lines, "\n")
}

func formatWordPressUpdatesDiscord(updates []checker.Update) string {
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
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}
