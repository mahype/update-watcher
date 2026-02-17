package webproject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mahype/update-watcher/checker"
)

func init() {
	RegisterManager(&YarnManager{})
}

// YarnManager checks for outdated yarn packages.
type YarnManager struct{}

func (y *YarnManager) Name() string { return "yarn" }

func (y *YarnManager) MarkerFiles() []string {
	return []string{"yarn.lock"}
}

func (y *YarnManager) CheckOutdated(project ProjectConfig) ([]checker.Update, error) {
	if y.isBerry(project) {
		return y.checkOutdatedBerry(project)
	}
	return y.checkOutdatedClassic(project)
}

// isBerry detects yarn v2+ (Berry) by checking the version.
func (y *YarnManager) isBerry(project ProjectConfig) bool {
	spec := BuildManagerCommand(project, "yarn", "--version")
	result, err := ExecuteCommand(spec)
	if err != nil || result == nil {
		return false
	}
	version := strings.TrimSpace(result.Stdout)
	return !strings.HasPrefix(version, "1.")
}

// yarnOutdatedTable is the NDJSON table structure from yarn v1 `yarn outdated --json`.
type yarnOutdatedTable struct {
	Type string `json:"type"`
	Data struct {
		Head []string   `json:"head"`
		Body [][]string `json:"body"`
	} `json:"data"`
}

// checkOutdatedClassic handles yarn v1.x.
func (y *YarnManager) checkOutdatedClassic(project ProjectConfig) ([]checker.Update, error) {
	spec := BuildManagerCommand(project, "yarn", "outdated", "--json")
	result, err := ExecuteCommand(spec)
	// yarn outdated exits 1 when packages are outdated
	if err != nil && result == nil {
		return nil, fmt.Errorf("yarn outdated failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	// yarn v1 outputs NDJSON - look for the "table" type line
	source := fmt.Sprintf("%s/yarn", project.Name)
	var updates []checker.Update

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		var table yarnOutdatedTable
		if err := json.Unmarshal([]byte(line), &table); err != nil {
			continue
		}
		if table.Type != "table" {
			continue
		}

		// Find column indices
		nameIdx, currentIdx, latestIdx := -1, -1, -1
		for i, h := range table.Data.Head {
			switch strings.ToLower(h) {
			case "package":
				nameIdx = i
			case "current":
				currentIdx = i
			case "latest":
				latestIdx = i
			}
		}

		if nameIdx < 0 || currentIdx < 0 || latestIdx < 0 {
			continue
		}

		for _, row := range table.Data.Body {
			if len(row) <= latestIdx {
				continue
			}
			name := row[nameIdx]
			current := row[currentIdx]
			latest := row[latestIdx]
			if current == latest {
				continue
			}
			updates = append(updates, checker.Update{
				Name:           name,
				CurrentVersion: current,
				NewVersion:     latest,
				Type:           checker.UpdateTypeRegular,
				Priority:       checker.PriorityNormal,
				Source:         source,
			})
		}
	}

	slog.Info("yarn outdated check complete", "project", project.Name, "updates", len(updates))
	return updates, nil
}

// checkOutdatedBerry handles yarn v2+ (Berry).
// Yarn Berry doesn't have a built-in `outdated` command by default.
// We use `yarn outdated` which requires the outdated plugin, or fall back
// to parsing `yarn npm info` per package.
func (y *YarnManager) checkOutdatedBerry(project ProjectConfig) ([]checker.Update, error) {
	// Try yarn outdated (requires the upgrade-interactive plugin)
	spec := BuildManagerCommand(project, "yarn", "outdated", "--json")
	result, err := ExecuteCommand(spec)
	if err != nil && result == nil {
		// Fallback: yarn Berry without outdated plugin - skip gracefully
		slog.Warn("yarn berry outdated not available (install plugin with: yarn plugin import outdated)",
			"project", project.Name)
		return nil, nil
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	// yarn berry outdated --json outputs one JSON object per line
	source := fmt.Sprintf("%s/yarn", project.Name)
	var updates []checker.Update

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		var entry struct {
			Name    string `json:"name"`
			Current string `json:"current"`
			Latest  string `json:"latest"`
			Type    string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Name == "" || entry.Current == entry.Latest {
			continue
		}
		updates = append(updates, checker.Update{
			Name:           entry.Name,
			CurrentVersion: entry.Current,
			NewVersion:     entry.Latest,
			Type:           checker.UpdateTypeRegular,
			Priority:       checker.PriorityNormal,
			Source:         source,
		})
	}

	slog.Info("yarn berry outdated check complete", "project", project.Name, "updates", len(updates))
	return updates, nil
}

// yarnAuditAdvisory represents an advisory from yarn v1 `yarn audit --json`.
type yarnAuditAdvisory struct {
	Type string `json:"type"`
	Data struct {
		Advisory struct {
			ModuleName string `json:"module_name"`
			Severity   string `json:"severity"`
			Title      string `json:"title"`
		} `json:"advisory"`
	} `json:"data"`
}

// Audit runs yarn audit and returns security updates.
func (y *YarnManager) Audit(project ProjectConfig) ([]checker.Update, error) {
	var spec CommandSpec
	if y.isBerry(project) {
		spec = BuildManagerCommand(project, "yarn", "npm", "audit", "--json")
	} else {
		spec = BuildManagerCommand(project, "yarn", "audit", "--json")
	}

	result, err := ExecuteCommand(spec)
	if err != nil && result == nil {
		return nil, fmt.Errorf("yarn audit failed: %w", err)
	}

	if result == nil || result.Stdout == "" {
		return nil, nil
	}

	source := fmt.Sprintf("%s/yarn", project.Name)
	seen := make(map[string]bool)
	var updates []checker.Update

	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		var advisory yarnAuditAdvisory
		if err := json.Unmarshal([]byte(line), &advisory); err != nil {
			continue
		}
		if advisory.Type != "auditAdvisory" {
			continue
		}
		name := advisory.Data.Advisory.ModuleName
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		updates = append(updates, checker.Update{
			Name:     name,
			Type:     checker.UpdateTypeSecurity,
			Priority: mapSeverityToPriority(advisory.Data.Advisory.Severity),
			Source:   source,
		})
	}

	slog.Info("yarn audit complete", "project", project.Name, "vulnerabilities", len(updates))
	return updates, nil
}
