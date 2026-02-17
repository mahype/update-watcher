package formatting

import (
	"fmt"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// Summary holds aggregated stats for a set of check results.
type Summary struct {
	TotalUpdates  int
	SecurityCount int
	CheckerCount  int
}

// SummarizeResults calculates aggregated stats from check results.
func SummarizeResults(results []*checker.CheckResult) Summary {
	s := Summary{CheckerCount: len(results)}
	for _, r := range results {
		s.TotalUpdates += len(r.Updates)
		for _, u := range r.Updates {
			if u.Type == checker.UpdateTypeSecurity {
				s.SecurityCount++
			}
		}
	}
	return s
}

// CheckerEmoji returns an emoji for a checker type.
func CheckerEmoji(name string, useEmoji bool) string {
	if !useEmoji {
		return ""
	}
	switch name {
	case "apt", "dnf", "pacman", "zypper", "apk":
		return "\U0001f427" // 🐧
	case "macos":
		return "\U0001f34e" // 🍎
	case "docker":
		return "\U0001f433" // 🐳
	case "wordpress":
		return "\U0001f4dd" // 📝
	case "webproject":
		return "\U0001f4e6" // 📦
	default:
		return "\U0001f504" // 🔄
	}
}

// CheckerDisplayName returns a human-readable name for a checker type.
func CheckerDisplayName(name string) string {
	switch name {
	case "apt":
		return "APT Updates"
	case "dnf":
		return "DNF Updates"
	case "pacman":
		return "Pacman Updates"
	case "zypper":
		return "Zypper Updates"
	case "apk":
		return "APK Updates"
	case "macos":
		return "macOS Updates"
	case "docker":
		return "Docker Updates"
	case "wordpress":
		return "WordPress Updates"
	case "webproject":
		return "Web Project Updates"
	default:
		return name + " Updates"
	}
}

// PriorityIndicator returns a visual indicator for an update's priority.
func PriorityIndicator(u checker.Update, useEmoji bool) string {
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

// FormatUpdatesMarkdown formats updates as markdown lines.
func FormatUpdatesMarkdown(r *checker.CheckResult, useEmoji bool) string {
	if len(r.Updates) == 0 {
		return ""
	}

	if r.CheckerName == "wordpress" {
		return formatWordPressUpdatesMarkdown(r.Updates, useEmoji)
	}

	if r.CheckerName == "webproject" {
		return formatWebProjectUpdatesMarkdown(r.Updates, useEmoji)
	}

	var lines []string
	for _, u := range r.Updates {
		indicator := PriorityIndicator(u, useEmoji)
		var line string
		if u.Type == checker.UpdateTypeSecurity {
			line = fmt.Sprintf("%s **`%s`** %s \u2192 %s \u26a0\ufe0f **SECURITY**", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		} else {
			line = fmt.Sprintf("%s `%s` %s \u2192 %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", u.Source)
		}
		if u.Phasing != "" {
			line += fmt.Sprintf(" _(phased %s)_", u.Phasing)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWordPressUpdatesMarkdown(updates []checker.Update, useEmoji bool) string {
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
		for _, u := range siteUpdates {
			indicator := PriorityIndicator(u, useEmoji)
			typeName := strings.ToUpper(u.Type[:1]) + u.Type[1:]
			var line string
			if u.Type == checker.UpdateTypeSecurity {
				line = fmt.Sprintf("%s %s: **`%s`** %s \u2192 %s \u26a0\ufe0f **SECURITY**", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			} else {
				line = fmt.Sprintf("%s %s: `%s` %s \u2192 %s", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			}
			lines = append(lines, line)
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

// FormatUpdatesPlainText formats updates as plain text lines.
func FormatUpdatesPlainText(r *checker.CheckResult) string {
	if len(r.Updates) == 0 {
		return ""
	}

	if r.CheckerName == "wordpress" {
		return formatWordPressUpdatesPlainText(r.Updates)
	}

	if r.CheckerName == "webproject" {
		return formatWebProjectUpdatesPlainText(r.Updates)
	}

	var lines []string
	for _, u := range r.Updates {
		indicator := "[!]"
		if u.Type != checker.UpdateTypeSecurity && u.Priority != checker.PriorityCritical {
			indicator = "[-]"
		}
		line := fmt.Sprintf("  %s %s %s -> %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
		if u.Type == checker.UpdateTypeSecurity {
			line += " [SECURITY]"
		}
		if u.Source != "" {
			line += fmt.Sprintf(" (%s)", u.Source)
		}
		if u.Phasing != "" {
			line += fmt.Sprintf(" [phased %s]", u.Phasing)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWebProjectUpdatesMarkdown(updates []checker.Update, useEmoji bool) string {
	type projectGroup struct {
		managers map[string][]checker.Update
		order    []string
	}
	projects := make(map[string]*projectGroup)
	var projectOrder []string

	for _, u := range updates {
		projectName, managerName := splitWebProjectSource(u.Source)
		if _, exists := projects[projectName]; !exists {
			projects[projectName] = &projectGroup{
				managers: make(map[string][]checker.Update),
			}
			projectOrder = append(projectOrder, projectName)
		}
		pg := projects[projectName]
		if _, exists := pg.managers[managerName]; !exists {
			pg.order = append(pg.order, managerName)
		}
		pg.managers[managerName] = append(pg.managers[managerName], u)
	}

	var sections []string
	for _, projectName := range projectOrder {
		pg := projects[projectName]
		lines := []string{fmt.Sprintf("**%s**", projectName)}
		for _, managerName := range pg.order {
			lines = append(lines, fmt.Sprintf("_%s:_", managerName))
			for _, u := range pg.managers[managerName] {
				indicator := PriorityIndicator(u, useEmoji)
				var line string
				if u.Type == checker.UpdateTypeSecurity {
					if u.CurrentVersion != "" && u.NewVersion != "" {
						line = fmt.Sprintf("%s **`%s`** %s \u2192 %s \u26a0\ufe0f **SECURITY**", indicator, u.Name, u.CurrentVersion, u.NewVersion)
					} else {
						line = fmt.Sprintf("%s **`%s`** \u26a0\ufe0f **SECURITY**", indicator, u.Name)
					}
				} else {
					line = fmt.Sprintf("%s `%s` %s \u2192 %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
				}
				lines = append(lines, line)
			}
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func formatWebProjectUpdatesPlainText(updates []checker.Update) string {
	type projectGroup struct {
		managers map[string][]checker.Update
		order    []string
	}
	projects := make(map[string]*projectGroup)
	var projectOrder []string

	for _, u := range updates {
		projectName, managerName := splitWebProjectSource(u.Source)
		if _, exists := projects[projectName]; !exists {
			projects[projectName] = &projectGroup{
				managers: make(map[string][]checker.Update),
			}
			projectOrder = append(projectOrder, projectName)
		}
		pg := projects[projectName]
		if _, exists := pg.managers[managerName]; !exists {
			pg.order = append(pg.order, managerName)
		}
		pg.managers[managerName] = append(pg.managers[managerName], u)
	}

	var sections []string
	for _, projectName := range projectOrder {
		pg := projects[projectName]
		lines := []string{fmt.Sprintf("  %s:", projectName)}
		for _, managerName := range pg.order {
			lines = append(lines, fmt.Sprintf("    %s:", managerName))
			for _, u := range pg.managers[managerName] {
				indicator := "[!]"
				if u.Type != checker.UpdateTypeSecurity && u.Priority != checker.PriorityCritical {
					indicator = "[-]"
				}
				var line string
				if u.CurrentVersion != "" && u.NewVersion != "" {
					line = fmt.Sprintf("      %s %s %s -> %s", indicator, u.Name, u.CurrentVersion, u.NewVersion)
				} else {
					line = fmt.Sprintf("      %s %s", indicator, u.Name)
				}
				if u.Type == checker.UpdateTypeSecurity {
					line += " [SECURITY]"
				}
				lines = append(lines, line)
			}
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

func splitWebProjectSource(source string) (string, string) {
	parts := strings.SplitN(source, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return source, ""
}

func formatWordPressUpdatesPlainText(updates []checker.Update) string {
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
		lines := []string{fmt.Sprintf("  %s:", source)}
		for _, u := range siteUpdates {
			indicator := "[!]"
			if u.Type != checker.UpdateTypeSecurity && u.Priority != checker.PriorityCritical {
				indicator = "[-]"
			}
			typeName := strings.ToUpper(u.Type[:1]) + u.Type[1:]
			line := fmt.Sprintf("    %s %s: %s %s -> %s", indicator, typeName, u.Name, u.CurrentVersion, u.NewVersion)
			if u.Type == checker.UpdateTypeSecurity {
				line += " [SECURITY]"
			}
			lines = append(lines, line)
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}
