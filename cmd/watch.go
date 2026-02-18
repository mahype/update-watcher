package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Add a watcher to the configuration",
	Long:  "Add a new update checker (apt, dnf, pacman, zypper, apk, docker, wordpress, webproject, or macos) to the configuration.",
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

// addWatchCommand creates and registers a watch subcommand using shared boilerplate.
// Use this for watchers that simply create a WatcherConfig and add it to the config file.
// Complex watchers (wordpress, webproject) that modify existing config entries should be
// implemented manually.
func addWatchCommand(use, short, displayName string, flagSetup func(*cobra.Command), buildConfig func(*cobra.Command) (config.WatcherConfig, error)) {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = config.NewDefault()
			}

			watcher, err := buildConfig(cmd)
			if err != nil {
				return err
			}

			cfg.AddWatcher(watcher)

			if err := config.Save(cfg, config.ConfigPath()); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("%s watcher added successfully.\n", displayName)
			return nil
		},
	}
	if flagSetup != nil {
		flagSetup(cmd)
	}
	watchCmd.AddCommand(cmd)
}
