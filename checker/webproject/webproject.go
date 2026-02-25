package webproject

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
)

func init() {
	checker.Register("webproject", NewFromConfig)
}

// ProjectConfig represents a single web project to check.
type ProjectConfig struct {
	Name        string
	Path        string
	RunAs       string
	Environment Environment
	Managers    []string // explicit list; empty = auto-detect
	CheckAudit  bool
}

// WebProjectChecker checks for outdated packages across web projects.
type WebProjectChecker struct {
	projects   []ProjectConfig
	checkAudit bool
}

// NewFromConfig creates a WebProjectChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	projects := parseProjects(cfg)
	if len(projects) == 0 {
		return nil, fmt.Errorf("no web projects configured")
	}
	return &WebProjectChecker{
		projects:   projects,
		checkAudit: cfg.GetBool("check_audit", true),
	}, nil
}

func parseProjects(cfg config.WatcherConfig) []ProjectConfig {
	raw := cfg.GetMapSlice("projects")
	var projects []ProjectConfig
	for _, m := range raw {
		pc := ProjectConfig{
			Environment: EnvAuto,
			CheckAudit:  cfg.GetBool("check_audit", true),
		}
		if v, ok := m["name"].(string); ok {
			pc.Name = v
		}
		if v, ok := m["path"].(string); ok {
			pc.Path = v
		}
		if v, ok := m["run_as"].(string); ok {
			pc.RunAs = v
		}
		if v, ok := m["environment"].(string); ok && v != "" {
			pc.Environment = Environment(v)
		}
		if v, ok := m["check_audit"].(bool); ok {
			pc.CheckAudit = v
		}
		// Parse explicit managers list
		if v, ok := m["managers"]; ok {
			switch val := v.(type) {
			case []interface{}:
				for _, item := range val {
					if s, ok := item.(string); ok {
						pc.Managers = append(pc.Managers, s)
					}
				}
			case []string:
				pc.Managers = val
			}
		}
		if pc.Path != "" {
			if pc.Name == "" {
				pc.Name = pc.Path
			}
			if pc.Environment == EnvAuto || pc.Environment == "" {
				pc.Environment = DetectEnvironment(pc.Path)
				slog.Info("auto-detected webproject environment",
					"project", pc.Name, "env", pc.Environment)
			}
			projects = append(projects, pc)
		}
	}
	return projects
}

func (w *WebProjectChecker) Name() string { return "webproject" }

func (w *WebProjectChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: w.Name(),
		CheckedAt:   time.Now(),
	}

	var allErrors []string

	for _, project := range w.projects {
		slog.Info("checking web project",
			"name", project.Name, "path", project.Path, "env", project.Environment)

		// Determine which package managers to check
		var managers []PackageManager
		if len(project.Managers) > 0 {
			for _, name := range project.Managers {
				if mgr, ok := GetManager(name); ok {
					managers = append(managers, mgr)
				} else {
					allErrors = append(allErrors,
						fmt.Sprintf("%s: unknown package manager %q", project.Name, name))
				}
			}
		} else {
			managers = DetectManagers(project.Path)
		}

		if len(managers) == 0 {
			allErrors = append(allErrors,
				fmt.Sprintf("%s: no supported package managers detected", project.Name))
			continue
		}

		slog.Info("detected package managers",
			"project", project.Name,
			"managers", managerNames(managers))

		for _, mgr := range managers {
			updates, err := mgr.CheckOutdated(project)
			if err != nil {
				allErrors = append(allErrors,
					fmt.Sprintf("%s/%s: %s", project.Name, mgr.Name(), err))
				continue
			}
			result.Updates = append(result.Updates, updates...)

			// Run security audit if supported and enabled
			if project.CheckAudit && w.checkAudit {
				if auditor, ok := mgr.(SecurityAuditor); ok {
					auditUpdates, err := auditor.Audit(project)
					if err != nil {
						slog.Warn("security audit failed",
							"project", project.Name, "manager", mgr.Name(), "error", err)
					} else {
						result.Updates = mergeAuditResults(result.Updates, auditUpdates, project.Name, mgr.Name())
					}
				}
			}
		}
	}

	if len(allErrors) > 0 {
		result.Error = strings.Join(allErrors, "; ")
	}

	if len(result.Updates) == 0 {
		result.Summary = "all projects are up to date"
	} else {
		secCount := 0
		for _, u := range result.Updates {
			if u.Type == checker.UpdateTypeSecurity {
				secCount++
			}
		}
		if secCount > 0 {
			result.Summary = fmt.Sprintf("%d outdated packages (%d security) across %d projects",
				len(result.Updates), secCount, len(w.projects))
		} else {
			result.Summary = fmt.Sprintf("%d outdated packages across %d projects",
				len(result.Updates), len(w.projects))
		}
	}

	return result, nil
}

func managerNames(managers []PackageManager) []string {
	names := make([]string, len(managers))
	for i, m := range managers {
		names[i] = m.Name()
	}
	return names
}

// mergeAuditResults upgrades matching updates to security type or adds new entries.
func mergeAuditResults(existing []checker.Update, audit []checker.Update, projectName string, managerName string) []checker.Update {
	source := fmt.Sprintf("%s/%s", projectName, managerName)

	auditMap := make(map[string]checker.Update)
	for _, a := range audit {
		auditMap[a.Name] = a
	}

	for i, u := range existing {
		if u.Source == source {
			if auditEntry, found := auditMap[u.Name]; found {
				existing[i].Type = checker.UpdateTypeSecurity
				if priorityRank(auditEntry.Priority) > priorityRank(existing[i].Priority) {
					existing[i].Priority = auditEntry.Priority
				}
				delete(auditMap, u.Name)
			}
		}
	}

	// Add audit-only entries not already in outdated list; skip those without a fix version
	for _, a := range auditMap {
		if a.NewVersion == "" {
			slog.Debug("skipping security entry without fix version", "package", a.Name, "source", a.Source)
			continue
		}
		existing = append(existing, a)
	}

	return existing
}
