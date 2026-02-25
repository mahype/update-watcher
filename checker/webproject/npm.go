package webproject

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mahype/update-watcher/checker"
)

func init() {
	RegisterManager(&NpmManager{})
}

// NpmManager checks for outdated npm packages.
type NpmManager struct{}

func (n *NpmManager) Name() string { return "npm" }

func (n *NpmManager) MarkerFiles() []string {
	return []string{"package-lock.json"}
}

// npmOutdatedEntry is the JSON structure from `npm outdated --json`.
type npmOutdatedEntry struct {
	Current  string `json:"current"`
	Wanted   string `json:"wanted"`
	Latest   string `json:"latest"`
	Type     string `json:"type"`
	Homepage string `json:"homepage"`
}

func (n *NpmManager) CheckOutdated(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "npm", "outdated", "--json")

	result, err := ExecuteCommand(spec)
	// npm outdated exits with code 1 when outdated packages exist
	if err != nil && result == nil {
		return nil, fmt.Errorf("npm outdated failed: %w", err)
	}

	if result == nil || result.Stdout == "" || result.Stdout == "{}" || result.Stdout == "{}\n" {
		return nil, nil
	}

	var outdated map[string]npmOutdatedEntry
	if err := json.Unmarshal([]byte(result.Stdout), &outdated); err != nil {
		return nil, fmt.Errorf("failed to parse npm outdated output: %w", err)
	}

	source := fmt.Sprintf("%s/npm", project.Name)
	var updates []checker.Update
	for name, entry := range outdated {
		if entry.Current == entry.Latest {
			continue
		}
		updates = append(updates, checker.Update{
			Name:           name,
			CurrentVersion: entry.Current,
			NewVersion:     entry.Latest,
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
			Source:         source,
		})
	}

	slog.Info("npm outdated check complete", "project", project.Name, "updates", len(updates))
	return updates, nil
}

// npmFixAvailable represents the fixAvailable field from npm audit, which is either
// false (no fix) or an object containing the fix version.
type npmFixAvailable struct {
	Version string
}

func (f *npmFixAvailable) UnmarshalJSON(data []byte) error {
	// false = no fix available
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		return nil
	}
	// object = fix available with version info
	var obj struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		f.Version = obj.Version
	}
	return nil
}

// npmAuditVulnerability represents a vulnerability from `npm audit --json`.
type npmAuditVulnerability struct {
	Name         string          `json:"name"`
	Severity     string          `json:"severity"`
	Range        string          `json:"range"`
	FixAvailable npmFixAvailable `json:"fixAvailable"`
}

type npmAuditOutput struct {
	Vulnerabilities map[string]npmAuditVulnerability `json:"vulnerabilities"`
}

// Audit runs npm audit and returns security updates.
func (n *NpmManager) Audit(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "npm", "audit", "--json")
	result, err := ExecuteCommand(spec)
	// npm audit exits non-zero when vulnerabilities found
	if err != nil && result == nil {
		return nil, fmt.Errorf("npm audit failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	var audit npmAuditOutput
	if err := json.Unmarshal([]byte(result.Stdout), &audit); err != nil {
		return nil, fmt.Errorf("failed to parse npm audit output: %w", err)
	}

	source := fmt.Sprintf("%s/npm", project.Name)
	var updates []checker.Update
	for name, vuln := range audit.Vulnerabilities {
		updates = append(updates, checker.Update{
			Name:       name,
			NewVersion: vuln.FixAvailable.Version,
			Type:       checker.UpdateTypeSecurity,
			Priority:   mapSeverityToPriority(vuln.Severity),
			Source:     source,
		})
	}

	slog.Info("npm audit complete", "project", project.Name, "vulnerabilities", len(updates))
	return updates, nil
}

func mapSeverityToPriority(severity string) string {
	switch severity {
	case "critical":
		return checker.PriorityCritical
	case "high":
		return checker.PriorityHigh
	case "moderate":
		return checker.PriorityNormal
	case "low", "info":
		return checker.PriorityLow
	default:
		return checker.PriorityNormal
	}
}
