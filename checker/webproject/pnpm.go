package webproject

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mahype/update-watcher/checker"
)

func init() {
	RegisterManager(&PnpmManager{})
}

// PnpmManager checks for outdated pnpm packages.
type PnpmManager struct{}

func (p *PnpmManager) Name() string { return "pnpm" }

func (p *PnpmManager) MarkerFiles() []string {
	return []string{"pnpm-lock.yaml"}
}

// pnpmOutdatedEntry is the JSON structure from `pnpm outdated --format json`.
type pnpmOutdatedEntry struct {
	Current        string `json:"current"`
	Latest         string `json:"latest"`
	Wanted         string `json:"wanted"`
	IsDeprecated   bool   `json:"isDeprecated"`
	DependencyType string `json:"dependencyType"`
}

func (p *PnpmManager) CheckOutdated(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "pnpm", "outdated", "--format", "json")

	result, err := ExecuteCommand(spec)
	// pnpm outdated exits with code 1 when outdated packages exist
	if err != nil && result == nil {
		return nil, fmt.Errorf("pnpm outdated failed: %w", err)
	}

	if result == nil || result.Stdout == "" || result.Stdout == "{}" || result.Stdout == "{}\n" {
		return nil, nil
	}

	var outdated map[string]pnpmOutdatedEntry
	if err := json.Unmarshal([]byte(result.Stdout), &outdated); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm outdated output: %w", err)
	}

	source := fmt.Sprintf("%s/pnpm", project.Name)
	var updates []checker.Update
	for name, entry := range outdated {
		if entry.Current == entry.Latest {
			continue
		}
		priority := checker.PriorityNormal
		if entry.IsDeprecated {
			priority = checker.PriorityHigh
		}
		updates = append(updates, checker.Update{
			Name:           name,
			CurrentVersion: entry.Current,
			NewVersion:     entry.Latest,
			Type:           checker.UpdateTypeRegular,
			Priority:       priority,
			Source:         source,
		})
	}

	slog.Info("pnpm outdated check complete", "project", project.Name, "updates", len(updates))
	return updates, nil
}

// pnpmAuditOutput mirrors the npm audit JSON structure used by pnpm.
type pnpmAuditOutput struct {
	Advisories map[string]struct {
		ModuleName string `json:"module_name"`
		Severity   string `json:"severity"`
		Title      string `json:"title"`
	} `json:"advisories"`
}

// Audit runs pnpm audit and returns security updates.
func (p *PnpmManager) Audit(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "pnpm", "audit", "--json")
	result, err := ExecuteCommand(spec)
	// pnpm audit exits non-zero when vulnerabilities found
	if err != nil && result == nil {
		return nil, fmt.Errorf("pnpm audit failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	var audit pnpmAuditOutput
	if err := json.Unmarshal([]byte(result.Stdout), &audit); err != nil {
		return nil, fmt.Errorf("failed to parse pnpm audit output: %w", err)
	}

	source := fmt.Sprintf("%s/pnpm", project.Name)
	seen := make(map[string]bool)
	var updates []checker.Update
	for _, advisory := range audit.Advisories {
		name := advisory.ModuleName
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		updates = append(updates, checker.Update{
			Name:     name,
			Type:     checker.UpdateTypeSecurity,
			Priority: mapSeverityToPriority(advisory.Severity),
			Source:   source,
		})
	}

	slog.Info("pnpm audit complete", "project", project.Name, "vulnerabilities", len(updates))
	return updates, nil
}
