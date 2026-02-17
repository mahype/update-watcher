package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/checker/wordpress"
	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchWordPressCmd = &cobra.Command{
	Use:   "wordpress",
	Short: "Add WordPress site update watcher",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		runAs, _ := cmd.Flags().GetString("run-as")
		envFlag, _ := cmd.Flags().GetString("env")
		noCore, _ := cmd.Flags().GetBool("no-core")
		noPlugins, _ := cmd.Flags().GetBool("no-plugins")
		noThemes, _ := cmd.Flags().GetBool("no-themes")

		if path == "" {
			return fmt.Errorf("--path is required")
		}
		if name == "" {
			name = path
		}

		// Resolve environment
		env := wordpress.Environment(envFlag)
		if env == wordpress.EnvAuto || env == "" {
			env = wordpress.DetectEnvironment(path)
			fmt.Printf("Auto-detected environment: %s (%s)\n", env.Label(), wordpress.EnvironmentDescription(env))
		}

		site := map[string]interface{}{
			"name":        name,
			"path":        path,
			"environment": string(env),
		}

		// Only set run_as for native environments
		if env.NeedsRunAs() && runAs != "" {
			site["run_as"] = runAs
		}

		// Check if wordpress watcher exists, add site to it
		var found bool
		for i, w := range cfg.Watchers {
			if w.Type == "wordpress" {
				sites := w.GetMapSlice("sites")
				sites = append(sites, site)
				sitesIface := make([]interface{}, len(sites))
				for j, s := range sites {
					sitesIface[j] = s
				}
				if cfg.Watchers[i].Options == nil {
					cfg.Watchers[i].Options = make(map[string]interface{})
				}
				cfg.Watchers[i].Options["sites"] = sitesIface
				found = true
				break
			}
		}

		if !found {
			watcher := config.WatcherConfig{
				Type:    "wordpress",
				Enabled: true,
				Options: map[string]interface{}{
					"sites":         []interface{}{site},
					"check_core":    !noCore,
					"check_plugins": !noPlugins,
					"check_themes":  !noThemes,
				},
			}
			cfg.AddWatcher(watcher)
		}

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("WordPress watcher added for %q at %s (env: %s)\n", name, path, env.Label())
		return nil
	},
}

func init() {
	watchWordPressCmd.Flags().String("path", "", "path to WordPress installation (required)")
	watchWordPressCmd.Flags().String("name", "", "human-readable site name (defaults to path)")
	watchWordPressCmd.Flags().String("run-as", "www-data", "OS user to run WP-CLI as (native only)")
	watchWordPressCmd.Flags().String("env", "auto", "environment type: auto, native, ddev, lando, wp-env, docker-compose, bedrock, local, mamp, xampp, laragon, valet")
	watchWordPressCmd.Flags().Bool("no-core", false, "skip core update check")
	watchWordPressCmd.Flags().Bool("no-plugins", false, "skip plugin update check")
	watchWordPressCmd.Flags().Bool("no-themes", false, "skip theme update check")
	watchCmd.AddCommand(watchWordPressCmd)
}
