package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/rootcheck"
	"github.com/mahype/update-watcher/wizard"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long:  "Walk through an interactive setup to configure update-watcher.",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("as-service-user")
		rootcheck.WarnOrReExec(force)

		// Load existing config (or start fresh)
		cfg, err := config.Load()
		if err != nil {
			cfg = config.NewDefault()
		}

		// Run the menu-driven wizard
		cfg, err = wizard.Run(cfg)
		if err != nil {
			return err
		}

		// Save config
		cfgPath := config.ConfigPath()
		if err := config.Save(cfg, cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("\nConfiguration saved to %s\n", cfgPath)
		fmt.Println("\nSetup complete! Run 'update-watcher run' to check for updates.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
