package zypper

import (
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// zypper list-updates output format (table):
// S | Repository  | Name            | Current Version | Available Version | Arch
// v | repo-oss    | vim             | 9.0.1234        | 9.0.1500          | x86_64
//
// The header separator line is "---+---+---..."
// Data lines start with "v" (or other status indicators).

// parseListUpdates parses the output of "zypper list-updates" into Updates.
func parseListUpdates(output string, securityPkgs map[string]bool) []checker.Update {
	var updates []checker.Update

	inTable := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect table start (separator line)
		if strings.HasPrefix(line, "--") && strings.Contains(line, "+") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		// Parse table row: split by |
		fields := strings.Split(line, "|")
		if len(fields) < 6 {
			continue
		}

		name := strings.TrimSpace(fields[2])
		currentVersion := strings.TrimSpace(fields[3])
		newVersion := strings.TrimSpace(fields[4])

		if name == "" || newVersion == "" {
			continue
		}

		isSecurity := securityPkgs != nil && securityPkgs[name]

		updateType := checker.UpdateTypeRegular
		priority := checker.PriorityNormal
		if isSecurity {
			updateType = checker.UpdateTypeSecurity
			priority = checker.PriorityHigh
		}

		updates = append(updates, checker.Update{
			Name:           name,
			CurrentVersion: currentVersion,
			NewVersion:     newVersion,
			Type:           updateType,
			Priority:       priority,
		})
	}

	return updates
}

// parseSecurityPatches extracts package names from "zypper list-patches --category security".
// Output format is similar table:
// Repository  | Name         | Category | Severity  | Interactive | Status   | Summary
// repo-update | patch-12345  | security | important | ---         | needed   | Security update for openssl
func parseSecurityPatches(output string) map[string]bool {
	pkgs := make(map[string]bool)

	inTable := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "--") && strings.Contains(line, "+") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}

		name := strings.TrimSpace(fields[1])
		status := strings.TrimSpace(fields[5])

		// Only mark packages from needed patches
		if strings.EqualFold(status, "needed") && name != "" {
			pkgs[name] = true
		}
	}

	return pkgs
}
