package cmd

import (
	"errors"
	"fmt"
	"os"
	"syscall"

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
			if errors.Is(err, wizard.ErrSelfUpdated) {
				// Save config before re-exec
				cfgPath := config.ConfigPath()
				if saveErr := config.Save(cfg, cfgPath); saveErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save config: %s\n", saveErr)
				}

				// Re-exec the updated binary with the same arguments
				binary, execErr := os.Executable()
				if execErr != nil {
					return fmt.Errorf("failed to determine executable path: %w", execErr)
				}
				fmt.Println("\n  Restarting with new version...")
				return syscall.Exec(binary, os.Args, os.Environ())
			}
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
