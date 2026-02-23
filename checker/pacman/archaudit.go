package pacman

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/internal/executil"
)

// archAuditRe matches lines like: "curl is affected by CVE-2023-46218. High risk!"
var archAuditRe = regexp.MustCompile(`^(\S+) is affected by .+\. (\w+) risk!$`)

// archAuditAvailable reports whether the arch-audit tool is installed.
func archAuditAvailable() bool {
	_, err := exec.LookPath("arch-audit")
	return err == nil
}

// runArchAudit executes arch-audit --upgradable and returns a map of
// package names to their highest severity level.
func runArchAudit() (map[string]string, error) {
	result, err := executil.Run("arch-audit", "--upgradable")
	if err != nil {
		// Exit code 1 means no vulnerable upgradable packages found.
		if result != nil && result.ExitCode == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("arch-audit --upgradable failed: %w", err)
	}
	return parseArchAudit(result.Stdout), nil
}

// parseArchAudit parses the output of arch-audit into a map of
// package name → severity (Critical, High, Medium, Low).
// If a package appears multiple times, the highest severity wins.
func parseArchAudit(output string) map[string]string {
	vulns := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		matches := archAuditRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		pkg := matches[1]
		severity := matches[2]

		if existing, ok := vulns[pkg]; ok {
			if severityRank(severity) > severityRank(existing) {
				vulns[pkg] = severity
			}
		} else {
			vulns[pkg] = severity
		}
	}
	return vulns
}

// mapArchAuditSeverity maps an arch-audit severity string to a checker priority constant.
func mapArchAuditSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return checker.PriorityCritical
	case "high":
		return checker.PriorityHigh
	case "medium":
		return checker.PriorityNormal
	case "low":
		return checker.PriorityLow
	default:
		return checker.PriorityNormal
	}
}

// enrichWithArchAudit sets Type and Priority on updates that have known vulnerabilities.
func enrichWithArchAudit(updates []checker.Update, vulns map[string]string) []checker.Update {
	if len(vulns) == 0 {
		return updates
	}
	for i, u := range updates {
		if severity, ok := vulns[u.Name]; ok {
			updates[i].Type = checker.UpdateTypeSecurity
			updates[i].Priority = mapArchAuditSeverity(severity)
		}
	}
	return updates
}

// severityRank returns a numeric rank for an arch-audit severity string.
func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
