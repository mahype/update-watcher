package wordpress

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
	checker.Register("wordpress", NewFromConfig)
}

// SiteConfig represents a single WordPress site to check.
type SiteConfig struct {
	Name        string
	Path        string
	RunAs       string
	Environment Environment
}

// WordPressChecker checks for WordPress core, plugin, and theme updates.
type WordPressChecker struct {
	sites        []SiteConfig
	checkCore    bool
	checkPlugins bool
	checkThemes  bool
}

// NewFromConfig creates a WordPressChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	sites := parseSites(cfg.GetMapSlice("sites"))
	if len(sites) == 0 {
		return nil, fmt.Errorf("no WordPress sites configured")
	}

	return &WordPressChecker{
		sites:        sites,
		checkCore:    cfg.GetBool("check_core", true),
		checkPlugins: cfg.GetBool("check_plugins", true),
		checkThemes:  cfg.GetBool("check_themes", true),
	}, nil
}

func parseSites(raw []map[string]interface{}) []SiteConfig {
	var sites []SiteConfig
	for _, m := range raw {
		sc := SiteConfig{
			RunAs:       "www-data",
			Environment: EnvAuto,
		}
		if v, ok := m["name"].(string); ok {
			sc.Name = v
		}
		if v, ok := m["path"].(string); ok {
			sc.Path = v
		}
		if v, ok := m["run_as"].(string); ok {
			sc.RunAs = v
		}
		if v, ok := m["environment"].(string); ok && v != "" {
			sc.Environment = Environment(v)
		}
		if sc.Path != "" {
			if sc.Name == "" {
				sc.Name = sc.Path
			}
			// Auto-detect environment if set to "auto"
			if sc.Environment == EnvAuto {
				sc.Environment = DetectEnvironment(sc.Path)
				slog.Info("auto-detected WordPress environment",
					"site", sc.Name, "env", sc.Environment)
			}
			sites = append(sites, sc)
		}
	}
	return sites
}

func (w *WordPressChecker) Name() string { return "wordpress" }

func (w *WordPressChecker) Check(ctx context.Context) (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: w.Name(),
		CheckedAt:   time.Now(),
	}

	var allErrors []string

	for _, site := range w.sites {
		slog.Info("checking WordPress site",
			"name", site.Name, "path", site.Path, "env", site.Environment)

		projectDir := site.Path
		if site.Environment.IsContainerBased() {
			projectDir = FindProjectDir(site.Path, site.Environment)
		}

		cli := &WPCLIRunner{
			Path:        site.Path,
			RunAs:       site.RunAs,
			Environment: site.Environment,
			ProjectDir:  projectDir,
		}

		if w.checkCore {
			updates, currentVersion, err := cli.CheckCoreUpdates()
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("%s core: %s", site.Name, err))
			} else {
				for _, u := range updates {
					priority := checker.PriorityNormal
					updateType := checker.UpdateTypeCore
					if u.Update == "minor" || u.Update == "patch" {
						priority = checker.PriorityHigh // WordPress security fixes are often minor releases
					}
					result.Updates = append(result.Updates, checker.Update{
						Name:           "core",
						CurrentVersion: strings.TrimSpace(currentVersion),
						NewVersion:     u.Version,
						Type:           updateType,
						Priority:       priority,
						Source:         site.Name,
					})
				}
			}
		}

		if w.checkPlugins {
			plugins, err := cli.CheckPluginUpdates()
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("%s plugins: %s", site.Name, err))
			} else {
				for _, p := range plugins {
					result.Updates = append(result.Updates, checker.Update{
						Name:           p.Name,
						CurrentVersion: p.Version,
						NewVersion:     p.UpdateVer,
						Type:           checker.UpdateTypePlugin,
						Priority:       checker.PriorityNormal,
						Source:         site.Name,
					})
				}
			}
		}

		if w.checkThemes {
			themes, err := cli.CheckThemeUpdates()
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("%s themes: %s", site.Name, err))
			} else {
				for _, t := range themes {
					result.Updates = append(result.Updates, checker.Update{
						Name:           t.Name,
						CurrentVersion: t.Version,
						NewVersion:     t.UpdateVer,
						Type:           checker.UpdateTypeTheme,
						Priority:       checker.PriorityLow,
						Source:         site.Name,
					})
				}
			}
		}
	}

	if len(allErrors) > 0 {
		result.Error = strings.Join(allErrors, "; ")
	}

	if len(result.Updates) == 0 {
		result.Summary = "all sites are up to date"
	} else {
		result.Summary = fmt.Sprintf("%d updates across %d sites", len(result.Updates), len(w.sites))
	}

	return result, nil
}
