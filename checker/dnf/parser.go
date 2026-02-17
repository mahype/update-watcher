package dnf

import (
	"regexp"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

// dnf check-update output format:
// package-name.arch                version                repository
// Example:
// vim-enhanced.x86_64              9.0.2136-1.fc39        updates
var checkUpdateRe = regexp.MustCompile(
	`^(\S+)\.(\S+)\s+(\S+)\s+(\S+)\s*$`,
)

// parseCheckUpdate parses the output of "dnf check-update" into Updates.
func parseCheckUpdate(output string, securityPkgs map[string]bool) []checker.Update {
	var updates []checker.Update

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip header lines like "Last metadata expiration check..."
		if strings.HasPrefix(line, "Last metadata") || strings.HasPrefix(line, "Obsoleting") {
			continue
		}

		matches := checkUpdateRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		pkgName := matches[1]
		newVersion := matches[3]

		isSecurity := securityPkgs != nil && securityPkgs[pkgName]

		updateType := checker.UpdateTypeRegular
		priority := checker.PriorityNormal
		if isSecurity {
			updateType = checker.UpdateTypeSecurity
			priority = checker.PriorityHigh
		}

		updates = append(updates, checker.Update{
			Name:       pkgName,
			NewVersion: newVersion,
			Type:       updateType,
			Priority:   priority,
		})
	}

	return updates
}

// dnf updateinfo list --security output format:
// FEDORA-2024-abc123  Important/Sec.  vim-enhanced-9.0.2136-1.fc39.x86_64
var securityInfoRe = regexp.MustCompile(
	`^\S+\s+\S+\s+(\S+?)(?:-\d)`,
)

// parseSecurityInfo extracts package names from "dnf updateinfo list --security" output.
func parseSecurityInfo(output string) map[string]bool {
	pkgs := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Extract the package name (last column, strip version-release.arch)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pkg := fields[len(fields)-1]
		// Strip .arch suffix
		if idx := strings.LastIndex(pkg, "."); idx > 0 {
			pkg = pkg[:idx]
		}
		// Strip version-release: find the last two hyphens (name-version-release)
		parts := strings.Split(pkg, "-")
		if len(parts) >= 3 {
			// Rejoin everything except last two parts (version-release)
			pkgName := strings.Join(parts[:len(parts)-2], "-")
			pkgs[pkgName] = true
		}
	}
	return pkgs
}

// parseSecurityUpdates parses security-only updates from "dnf updateinfo list --security".
func parseSecurityUpdates(output string) []checker.Update {
	var updates []checker.Update
	seen := make(map[string]bool)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pkg := fields[len(fields)-1]

		// Strip .arch suffix
		nameWithVersion := pkg
		if idx := strings.LastIndex(nameWithVersion, "."); idx > 0 {
			nameWithVersion = nameWithVersion[:idx]
		}

		// Extract name and version: name-version-release
		parts := strings.Split(nameWithVersion, "-")
		if len(parts) < 3 {
			continue
		}

		pkgName := strings.Join(parts[:len(parts)-2], "-")
		newVersion := strings.Join(parts[len(parts)-2:], "-")

		if seen[pkgName] {
			continue
		}
		seen[pkgName] = true

		updates = append(updates, checker.Update{
			Name:       pkgName,
			NewVersion: newVersion,
			Type:       checker.UpdateTypeSecurity,
			Priority:   checker.PriorityHigh,
		})
	}

	return updates
}
