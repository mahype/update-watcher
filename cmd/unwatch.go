package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var unwatchCmd = &cobra.Command{
	Use:   "unwatch <type>",
	Short: "Remove a watcher from the configuration",
	Long:  "Remove an update checker (apt, dnf, pacman, zypper, apk, docker, or wordpress) from the configuration.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		watcherType := args[0]
		name, _ := cmd.Flags().GetString("name")

		if !cfg.RemoveWatcher(watcherType, name) {
			return fmt.Errorf("no watcher of type %q found", watcherType)
		}

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Watcher %q removed successfully.\n", watcherType)
		return nil
	},
}

func init() {
	unwatchCmd.Flags().String("name", "", "for WordPress: remove a specific site by name")
	rootCmd.AddCommand(unwatchCmd)
}
