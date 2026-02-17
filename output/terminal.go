package output

import (
	"fmt"
	"strings"

	"github.com/mahype/update-watcher/checker"
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
		fmt.Printf("\n--- Errors ---\n")
		for _, err := range errors {
			fmt.Printf("  ! %s\n", err)
		}
	}

	fmt.Printf("\nTotal: %d updates found\n", totalUpdates)
}

func printCheckerResult(r *checker.CheckResult) {
	icon := checkerIcon(r.CheckerName)
	fmt.Printf("%s %s — %s\n", icon, strings.ToUpper(r.CheckerName), r.Summary)

	if r.Error != "" {
		fmt.Printf("  WARNING: %s\n", r.Error)
	}

	if r.CheckerName == "wordpress" {
		printWordPressUpdates(r.Updates)
	} else {
		for _, u := range r.Updates {
			marker := " "
			if u.Type == checker.UpdateTypeSecurity || u.Priority == checker.PriorityCritical {
				marker = "!"
			}
			if u.Type == checker.UpdateTypeSecurity {
				fmt.Printf("  [%s] %-30s %s -> %s  [SECURITY]\n", marker, u.Name, u.CurrentVersion, u.NewVersion)
			} else {
				fmt.Printf("  [%s] %-30s %s -> %s\n", marker, u.Name, u.CurrentVersion, u.NewVersion)
			}
		}
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

func checkerIcon(name string) string {
	switch name {
	case "apt":
		return "[APT]"
	case "docker":
		return "[DCK]"
	case "wordpress":
		return "[WP]"
	default:
		return "[???]"
	}
}
