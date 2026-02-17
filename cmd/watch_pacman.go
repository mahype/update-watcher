package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/spf13/cobra"
)

var watchPacmanCmd = &cobra.Command{
	Use:   "pacman",
	Short: "Add Pacman package update watcher (Arch/Manjaro)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		noSudo, _ := cmd.Flags().GetBool("no-sudo")

		watcher := config.WatcherConfig{
			Type:    "pacman",
			Enabled: true,
			Options: map[string]interface{}{
				"use_sudo": !noSudo,
			},
		}

		cfg.AddWatcher(watcher)

		if err := config.Save(cfg, config.ConfigPath()); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Pacman watcher added successfully.")
		return nil
	},
}

func init() {
	watchPacmanCmd.Flags().Bool("no-sudo", false, "do not use sudo for pacman sync operations")
	watchCmd.AddCommand(watchPacmanCmd)
}
