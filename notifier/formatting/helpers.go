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
	case "npm", "webproject", "snap", "flatpak":
		return "\U0001f4e6" // 📦
	case "homebrew":
		return "\U0001f37a" // 🍺
	case "openclaw":
		return "\U0001f43e" // 🐾
	case "distro":
		return "\U0001f4bf" // 💿
	case "self-update":
		return "\U0001f199" // 🆙
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
	case "npm":
		return "npm Global Updates"
	case "docker":
		return "Docker Updates"
	case "wordpress":
		return "WordPress Updates"
	case "webproject":
		return "Web Project Updates"
	case "openclaw":
		return "OpenClaw Updates"
	case "distro":
		return "Distro Release"
	case "self-update":
		return "Update Watcher"
	default:
		return name + " Updates"
	}
}

// UpdateCommand returns the shell command to apply updates for a given checker.
// Returns empty string if no single command is applicable.
func UpdateCommand(checkerName string) string {
	switch checkerName {
	case "apt":
		return "sudo apt upgrade"
	case "dnf":
		return "sudo dnf upgrade"
	case "pacman":
		return "sudo pacman -Syu"
	case "zypper":
		return "sudo zypper update"
	case "apk":
		return "sudo apk upgrade"
	case "macos":
		return "softwareupdate -i -a"
	case "homebrew":
		return "brew upgrade"
	case "npm":
		return "npm update -g"
	case "snap":
		return "sudo snap refresh"
	case "flatpak":
		return "flatpak update"
	case "docker":
		return "docker pull <image>"
	case "wordpress":
		return "wp plugin update --all && wp theme update --all && wp core update"
	case "openclaw":
		return "openclaw update"
	case "distro":
		return ""
	case "self-update":
		return "update-watcher self-update"
	default:
		return ""
	}
}

// PhasingNote checks if any updates are phased and returns the count
// and the command to force-install them. Returns 0 and empty string if
// no phased updates exist or phasing is not applicable for the checker.
// Held-back packages (Phasing == "held") are excluded from the count.
func PhasingNote(checkerName string, updates []checker.Update) (count int, command string) {
	for _, u := range updates {
		if u.Phasing != "" && u.Phasing != "held" {
			count++
		}
	}
	if count == 0 || checkerName != "apt" {
		return 0, ""
	}
	return count, "sudo apt-get -o APT::Get::Always-Include-Phased-Updates=true upgrade"
}

// KeptBackNote checks if any updates are held back and returns the count
// and the command to install them. Returns 0 and empty string if
// no held-back updates exist or the checker is not apt.
func KeptBackNote(checkerName string, updates []checker.Update) (count int, command string) {
	for _, u := range updates {
		if u.Phasing == "held" {
			count++
		}
	}
	if count == 0 || checkerName != "apt" {
		return 0, ""
	}
	return count, "sudo apt full-upgrade"
}

// UpdateCommandForResult returns the shell command to apply updates, taking
// held-back packages into account. For apt, if any update is held back,
// it returns "sudo apt full-upgrade" instead of "sudo apt upgrade".
func UpdateCommandForResult(checkerName string, updates []checker.Update) string {
	if checkerName == "apt" {
		for _, u := range updates {
			if u.Phasing == "held" {
				return "sudo apt full-upgrade"
			}
		}
	}
	return UpdateCommand(checkerName)
}

// UpdateGroup holds updates grouped by a key (e.g., site name, source).
type UpdateGroup struct {
	Key     string
	Updates []checker.Update
}

// GroupUpdatesBySource groups updates by their Source field, maintaining insertion order.
func GroupUpdatesBySource(updates []checker.Update) []UpdateGroup {
	grouped := make(map[string][]checker.Update)
	var order []string
	for _, u := range updates {
		if _, exists := grouped[u.Source]; !exists {
			order = append(order, u.Source)
		}
		grouped[u.Source] = append(grouped[u.Source], u)
	}
	result := make([]UpdateGroup, len(order))
	for i, key := range order {
		result[i] = UpdateGroup{Key: key, Updates: grouped[key]}
	}
	return result
}

// ProjectManagerGroup holds updates grouped by project and then by manager.
type ProjectManagerGroup struct {
	ProjectName string
	Managers    []UpdateGroup
}

// GroupByProjectAndManager groups web project updates by project name and manager.
func GroupByProjectAndManager(updates []checker.Update) []ProjectManagerGroup {
	type projectData struct {
		managers map[string][]checker.Update
		order    []string
	}
	projects := make(map[string]*projectData)
	var projectOrder []string

	for _, u := range updates {
		projectName, managerName := splitWebProjectSource(u.Source)
		if _, exists := projects[projectName]; !exists {
			projects[projectName] = &projectData{
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

	var result []ProjectManagerGroup
	for _, projectName := range projectOrder {
		pg := projects[projectName]
		var managers []UpdateGroup
		for _, managerName := range pg.order {
			managers = append(managers, UpdateGroup{Key: managerName, Updates: pg.managers[managerName]})
		}
		result = append(result, ProjectManagerGroup{ProjectName: projectName, Managers: managers})
	}
	return result
}

// PriorityIndicator returns a visual indicator for an update's priority.
func PriorityIndicator(u checker.Update, useEmoji bool) string {
	if !useEmoji {
		if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
			return "[!]"
		}
		if u.Phasing != "" {
			return "[~]"
		}
		return "[-]"
	}
	if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
		return "\U0001f534" // 🔴
	}
	if u.Phasing != "" {
		return "\U0001f7e1" // 🟡
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
		if u.Phasing == "held" {
			line += " _(kept back)_"
		} else if u.Phasing != "" {
			line += fmt.Sprintf(" _(phased %s)_", u.Phasing)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWordPressUpdatesMarkdown(updates []checker.Update, useEmoji bool) string {
	groups := GroupUpdatesBySource(updates)
	var sections []string
	for _, g := range groups {
		lines := []string{fmt.Sprintf("**%s**", g.Key)}
		for _, u := range g.Updates {
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
		if u.Phasing == "held" {
			line += " [kept back]"
		} else if u.Phasing != "" {
			line += fmt.Sprintf(" [phased %s]", u.Phasing)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func formatWebProjectUpdatesMarkdown(updates []checker.Update, useEmoji bool) string {
	groups := GroupByProjectAndManager(updates)
	var sections []string
	for _, pg := range groups {
		lines := []string{fmt.Sprintf("**%s**", pg.ProjectName)}
		for _, mg := range pg.Managers {
			lines = append(lines, fmt.Sprintf("_%s:_", mg.Key))
			for _, u := range mg.Updates {
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
	groups := GroupByProjectAndManager(updates)
	var sections []string
	for _, pg := range groups {
		lines := []string{fmt.Sprintf("  %s:", pg.ProjectName)}
		for _, mg := range pg.Managers {
			lines = append(lines, fmt.Sprintf("    %s:", mg.Key))
			for _, u := range mg.Updates {
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
	groups := GroupUpdatesBySource(updates)
	var sections []string
	for _, g := range groups {
		lines := []string{fmt.Sprintf("  %s:", g.Key)}
		for _, u := range g.Updates {
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
