package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchApkCmd = &cobra.Command{
	Use:   "apk",
	Short: "Add APK package update watcher (Alpine Linux)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		noSudo, _ := cmd.Flags().GetBool("no-sudo")

		watcher := config.WatcherConfig{
			Type:    "apk",
			Enabled: true,
			Options: map[string]interface{}{
				"use_sudo": !noSudo,
			},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("APK watcher added successfully.")
		return nil
	},
}

func init() {
	watchApkCmd.Flags().Bool("no-sudo", false, "do not use sudo for apk operations")
	watchCmd.AddCommand(watchApkCmd)
}
