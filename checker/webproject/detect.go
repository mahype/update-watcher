package webproject

import (
	"path/filepath"
)

// DetectManagers scans the project path for marker files and returns
// all matching package managers.
func DetectManagers(projectPath string) []PackageManager {
	var detected []PackageManager

	for _, mgr := range AllManagers() {
		for _, marker := range mgr.MarkerFiles() {
			markerPath := filepath.Join(projectPath, marker)
			if fileExists(markerPath) {
				detected = append(detected, mgr)
				break
			}
		}
	}

	return resolveNodeConflicts(detected, projectPath)
}

// resolveNodeConflicts handles cases where multiple Node.js managers are detected.
// When multiple Node.js lockfiles exist, only the most specific one is kept.
// Priority: pnpm > yarn > npm (based on lockfile presence).
func resolveNodeConflicts(managers []PackageManager, projectPath string) []PackageManager {
	hasNpm := false
	hasYarn := false
	hasPnpm := false

	for _, m := range managers {
		switch m.Name() {
		case "npm":
			hasNpm = true
		case "yarn":
			hasYarn = true
		case "pnpm":
			hasPnpm = true
		}
	}

	nodeCount := 0
	if hasNpm {
		nodeCount++
	}
	if hasYarn {
		nodeCount++
	}
	if hasPnpm {
		nodeCount++
	}

	// If only one or zero Node.js managers, no conflict to resolve
	if nodeCount <= 1 {
		return managers
	}

	// Resolve based on lockfile presence
	if hasPnpm && fileExists(filepath.Join(projectPath, "pnpm-lock.yaml")) {
		hasNpm = false
		hasYarn = false
	} else if hasYarn && fileExists(filepath.Join(projectPath, "yarn.lock")) {
		hasNpm = false
		hasPnpm = false
	}
	// If no lockfile resolves it, keep npm as default
	if hasNpm && hasYarn && hasPnpm {
		hasYarn = false
		hasPnpm = false
	}

	var result []PackageManager
	for _, m := range managers {
		switch m.Name() {
		case "npm":
			if hasNpm {
				result = append(result, m)
			}
		case "yarn":
			if hasYarn {
				result = append(result, m)
			}
		case "pnpm":
			if hasPnpm {
				result = append(result, m)
			}
		default:
			result = append(result, m) // non-Node managers always kept
		}
	}
	return result
}
