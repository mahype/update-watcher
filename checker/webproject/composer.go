package webproject

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

func init() {
	RegisterManager(&ComposerManager{})
}

// ComposerManager checks for outdated Composer packages.
type ComposerManager struct{}

func (c *ComposerManager) Name() string { return "composer" }

func (c *ComposerManager) MarkerFiles() []string {
	return []string{"composer.json"}
}

// composerOutdatedOutput is the JSON structure from `composer outdated --format=json --direct`.
type composerOutdatedOutput struct {
	Installed []composerPackage `json:"installed"`
}

// abandonedField handles Composer's "abandoned" field which can be either
// a boolean (true) or a string ("use/other-package").
type abandonedField string

func (a *abandonedField) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*a = abandonedField(s)
		return nil
	}
	// Try boolean (Composer returns true when abandoned without replacement)
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		if b {
			*a = "yes"
		}
		return nil
	}
	return fmt.Errorf("abandoned field: expected string or bool, got %s", string(data))
}

type composerPackage struct {
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Latest       string         `json:"latest"`
	LatestStatus string         `json:"latest-status"`
	Description  string         `json:"description"`
	Abandoned    abandonedField `json:"abandoned,omitempty"`
}

func (c *ComposerManager) CheckOutdated(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "composer", "outdated", "--format=json", "--direct")

	result, err := ExecuteCommand(spec)
	if err != nil && result == nil {
		return nil, fmt.Errorf("composer outdated failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	var output composerOutdatedOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		return nil, fmt.Errorf("failed to parse composer outdated output: %w", err)
	}

	source := fmt.Sprintf("%s/composer", project.Name)
	var updates []checker.Update
	for _, pkg := range output.Installed {
		if pkg.LatestStatus == "up-to-date" {
			continue
		}
		priority := checker.PriorityNormal
		if pkg.LatestStatus == "semver-safe-update" {
			priority = checker.PriorityLow
		}
		if string(pkg.Abandoned) != "" {
			priority = checker.PriorityHigh
		}
		updates = append(updates, checker.Update{
			Name:           pkg.Name,
			CurrentVersion: pkg.Version,
			NewVersion:     pkg.Latest,
			Type:           checker.UpdateTypeRegular,
			Priority:       priority,
			Source:         source,
		})
	}

	slog.Info("composer outdated check complete", "project", project.Name, "updates", len(updates))
	return updates, nil
}

// composerAuditOutput is the JSON structure from `composer audit --format=json`.
type composerAuditOutput struct {
	Advisories map[string][]composerAdvisory `json:"advisories"`
}

type composerAdvisory struct {
	AdvisoryID       string `json:"advisoryId"`
	PackageName      string `json:"packageName"`
	AffectedVersions string `json:"affectedVersions"`
	Title            string `json:"title"`
	Severity         string `json:"severity"`
}

// Audit runs composer audit and returns security updates.
func (c *ComposerManager) Audit(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "composer", "audit", "--format=json")
	result, err := ExecuteCommand(spec)
	// composer audit exits non-zero when vulnerabilities found
	if err != nil && result == nil {
		return nil, fmt.Errorf("composer audit failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	var audit composerAuditOutput
	if err := json.Unmarshal([]byte(result.Stdout), &audit); err != nil {
		return nil, fmt.Errorf("failed to parse composer audit output: %w", err)
	}

	source := fmt.Sprintf("%s/composer", project.Name)
	var updates []checker.Update
	for pkgName, advisories := range audit.Advisories {
		// Use the highest severity from all advisories for a package
		highestPriority := checker.PriorityLow
		for _, adv := range advisories {
			p := mapComposerSeverityToPriority(adv.Severity)
			if priorityRank(p) > priorityRank(highestPriority) {
				highestPriority = p
			}
		}
		// Best advisory: use the one with highest severity to extract fix version
		fixVersion := ""
		for _, adv := range advisories {
			if v := extractFixVersion(adv.AffectedVersions); v != "" {
				fixVersion = v
				break
			}
		}
		updates = append(updates, checker.Update{
			Name:       pkgName,
			NewVersion: fixVersion,
			Type:       checker.UpdateTypeSecurity,
			Priority:   highestPriority,
			Source:     source,
		})
	}

	slog.Info("composer audit complete", "project", project.Name, "vulnerabilities", len(updates))
	return updates, nil
}

// extractFixVersion extracts the minimum fix version from a composer AffectedVersions
// constraint string like ">=1.0,<1.5.2" or "<2.0.0". Returns empty string if no
// upper bound is found.
func extractFixVersion(affectedVersions string) string {
	for _, part := range strings.Split(affectedVersions, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "<") && !strings.HasPrefix(part, "<=") {
			return strings.TrimPrefix(part, "<")
		}
	}
	return ""
}

func mapComposerSeverityToPriority(severity string) string {
	switch severity {
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

func priorityRank(p string) int {
	switch p {
	case checker.PriorityCritical:
		return 4
	case checker.PriorityHigh:
		return 3
	case checker.PriorityNormal:
		return 2
	case checker.PriorityLow:
		return 1
	default:
		return 0
	}
}
