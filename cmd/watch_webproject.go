package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/checker/webproject"
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchWebProjectCmd = &cobra.Command{
	Use:   "webproject",
	Short: "Add web project dependency watcher",
	Long:  "Watch a web project for outdated npm, yarn, pnpm, or composer packages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		runAs, _ := cmd.Flags().GetString("run-as")
		envFlag, _ := cmd.Flags().GetString("env")
		managersFlag, _ := cmd.Flags().GetStringSlice("managers")
		noAudit, _ := cmd.Flags().GetBool("no-audit")

		if path == "" {
			return fmt.Errorf("--path is required")
		}
		if name == "" {
			name = path
		}

		// Resolve environment
		env := webproject.Environment(envFlag)
		if env == webproject.EnvAuto || env == "" {
			env = webproject.DetectEnvironment(path)
			fmt.Printf("Auto-detected environment: %s (%s)\n", env.Label(), webproject.EnvironmentDescription(env))
		}

		// Auto-detect managers if not specified
		if len(managersFlag) == 0 {
			detected := webproject.DetectManagers(path)
			for _, m := range detected {
				managersFlag = append(managersFlag, m.Name())
			}
			if len(managersFlag) > 0 {
				fmt.Printf("Detected package managers: %v\n", managersFlag)
			} else {
				return fmt.Errorf("no supported package managers found at %s", path)
			}
		}

		project := map[string]interface{}{
			"name":        name,
			"path":        path,
			"environment": string(env),
			"check_audit": !noAudit,
		}
		if runAs != "" {
			project["run_as"] = runAs
		}
		if len(managersFlag) > 0 {
			mgrs := make([]interface{}, len(managersFlag))
			for i, m := range managersFlag {
				mgrs[i] = m
			}
			project["managers"] = mgrs
		}

		// Check if webproject watcher exists, add project to it
		var found bool
		for i, w := range cfg.Watchers {
			if w.Type == "webproject" {
				projects := w.GetMapSlice("projects")
				projects = append(projects, project)
				projectsIface := make([]interface{}, len(projects))
				for j, p := range projects {
					projectsIface[j] = p
				}
				if cfg.Watchers[i].Options == nil {
					cfg.Watchers[i].Options = make(map[string]interface{})
				}
				cfg.Watchers[i].Options["projects"] = projectsIface
				found = true
				break
			}
		}

		if !found {
			watcher := config.WatcherConfig{
				Type:    "webproject",
				Enabled: true,
				Options: map[string]interface{}{
					"check_audit": !noAudit,
					"projects":    []interface{}{project},
				},
			}
			cfg.AddWatcher(watcher)
		}

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Web project watcher added for %q at %s (env: %s, managers: %v)\n",
			name, path, env.Label(), managersFlag)
		return nil
	},
}

func init() {
	watchWebProjectCmd.Flags().String("path", "", "path to web project root (required)")
	watchWebProjectCmd.Flags().String("name", "", "human-readable project name (defaults to path)")
	watchWebProjectCmd.Flags().String("run-as", "", "OS user for native execution")
	watchWebProjectCmd.Flags().String("env", "auto", "environment: auto, native, ddev, lando, docker-compose")
	watchWebProjectCmd.Flags().StringSlice("managers", nil, "package managers to check (auto-detect if omitted): npm, yarn, pnpm, composer")
	watchWebProjectCmd.Flags().Bool("no-audit", false, "skip security audit checks")
	watchCmd.AddCommand(watchWebProjectCmd)
}
