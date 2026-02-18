package output

import (
	"fmt"
	"strings"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/notifier/formatting"
)

// PrintResults prints check results to the terminal.
func PrintResults(results []*checker.CheckResult, errors []error) {
	totalUpdates := 0
	for _, r := range results {
		totalUpdates += len(r.Updates)
	}

	fmt.Printf("\n=== Update Report ===\n\n")

	for _, r := range results {
		printCheckerResult(r)
	}

	if len(errors) > 0 {
		fmt.Printf("\n\n--- Errors ---\n")
		for _, err := range errors {
			fmt.Printf("  ! %s\n", err)
		}
	}

	fmt.Printf("\nTotal: %d updates found\n", totalUpdates)
}

func printCheckerResult(r *checker.CheckResult) {
	icon := checkerIcon(r.CheckerName)
	fmt.Printf("%s %s — %s\n\n", icon, strings.ToUpper(r.CheckerName), r.Summary)

	if r.Error != "" {
		fmt.Printf("  WARNING: %s\n\n", r.Error)
	}

	if r.CheckerName == "wordpress" {
		printWordPressUpdates(r.Updates)
	} else if r.CheckerName == "webproject" {
		printWebProjectUpdates(r.Updates)
	} else {
		for _, u := range r.Updates {
			marker := " "
			if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
				marker = "!"
			}
			suffix := ""
			if u.Type == checker.UpdateTypeSecurity {
				suffix += "  [SECURITY]"
			}
			if u.Phasing != "" {
				suffix += fmt.Sprintf("  [phased %s]", u.Phasing)
			}
			fmt.Printf("  [%s] %-30s %s -> %s%s\n", marker, u.Name, u.CurrentVersion, u.NewVersion, suffix)
		}
	}

	if cmd := formatting.UpdateCommand(r.CheckerName); cmd != "" && len(r.Updates) > 0 {
		fmt.Printf("\n  \U0001f4a1 Update: %s\n", cmd)
	} else if len(r.Updates) > 0 && r.Updates[0].Source != "" {
		fmt.Printf("\n  \U0001f4a1 Update: %s\n", r.Updates[0].Source)
	}

	fmt.Println()
}

func printWordPressUpdates(updates []checker.Update) {
	grouped := make(map[string][]checker.Update)
	var order []string
	for _, u := range updates {
		if _, exists := grouped[u.Source]; !exists {
			order = append(order, u.Source)
		}
		grouped[u.Source] = append(grouped[u.Source], u)
	}

	for _, source := range order {
		fmt.Printf("  %s:\n", source)
		for _, u := range grouped[source] {
			marker := " "
			if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
				marker = "!"
			}
			if u.Type == checker.UpdateTypeSecurity {
				fmt.Printf("    [%s] %-10s %-25s %s -> %s  [SECURITY]\n", marker, u.Type, u.Name, u.CurrentVersion, u.NewVersion)
			} else {
				fmt.Printf("    [%s] %-10s %-25s %s -> %s\n", marker, u.Type, u.Name, u.CurrentVersion, u.NewVersion)
			}
		}
	}
}

func printWebProjectUpdates(updates []checker.Update) {
	// Group by project (first part of Source before "/") then by manager
	type projectGroup struct {
		managers map[string][]checker.Update
		order    []string
	}
	projects := make(map[string]*projectGroup)
	var projectOrder []string

	for _, u := range updates {
		projectName, managerName := splitSource(u.Source)
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

	for _, projectName := range projectOrder {
		pg := projects[projectName]
		fmt.Printf("  %s:\n", projectName)
		for _, managerName := range pg.order {
			fmt.Printf("    %s:\n", managerName)
			for _, u := range pg.managers[managerName] {
				marker := " "
				if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
					marker = "!"
				}
				suffix := ""
				if u.Type == checker.UpdateTypeSecurity {
					suffix = "  [SECURITY]"
				}
				if u.CurrentVersion != "" && u.NewVersion != "" {
					fmt.Printf("      [%s] %-30s %s -> %s%s\n", marker, u.Name, u.CurrentVersion, u.NewVersion, suffix)
				} else {
					fmt.Printf("      [%s] %-30s%s\n", marker, u.Name, suffix)
				}
			}
		}
	}
}

func splitSource(source string) (string, string) {
	parts := strings.SplitN(source, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return source, ""
}

func checkerIcon(name string) string {
	switch name {
	case "apt":
		return "[APT]"
	case "dnf":
		return "[DNF]"
	case "pacman":
		return "[PAC]"
	case "zypper":
		return "[ZYP]"
	case "apk":
		return "[APK]"
	case "macos":
		return "[MAC]"
	case "homebrew":
		return "[BREW]"
	case "snap":
		return "[SNAP]"
	case "flatpak":
		return "[FLAT]"
	case "docker":
		return "[DCK]"
	case "wordpress":
		return "[WP]"
	case "webproject":
		return "[WEB]"
	case "distro":
		return "[DST]"
	default:
		return "[???]"
	}
}
