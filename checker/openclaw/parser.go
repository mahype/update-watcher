package openclaw

import (
	"regexp"
	"strings"
)

// summaryRe matches the final summary line, e.g.:
// "Update available (npm 2026.2.17). Run: openclaw update"
var summaryRe = regexp.MustCompile(`Update available \(\S+\s+(\S+)\)`)

// tableUpdateRe matches the Update row in the status table, e.g.:
// "│ Update   │ available · pnpm · npm update 2026.2.17 │"
var tableUpdateRe = regexp.MustCompile(`Update\s*│\s*available\b.*?(\d[\d.]+\d)`)

// parseStatus extracts the new version and whether an update is available
// from the output of "openclaw update status".
func parseStatus(output string) (newVersion string, available bool) {
	// Try the summary line first (most reliable).
	if m := summaryRe.FindStringSubmatch(output); m != nil {
		return m[1], true
	}

	// Fall back to the table row.
	if m := tableUpdateRe.FindStringSubmatch(output); m != nil {
		return m[1], true
	}

	return "", false
}

// versionRe extracts a version number from openclaw --version output.
var versionRe = regexp.MustCompile(`(\d[\d.]+\d)`)

// parseVersion extracts the version string from "openclaw --version" output.
func parseVersion(output string) string {
	output = strings.TrimSpace(output)
	if m := versionRe.FindString(output); m != "" {
		return m
	}
	return output
}
