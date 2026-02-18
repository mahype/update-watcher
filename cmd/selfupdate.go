package cmd

import (
	"fmt"
	"os"

	"github.com/mahype/update-watcher/internal/selfupdate"
	"github.com/mahype/update-watcher/internal/version"
	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update update-watcher to the latest version",
	Long:  "Download and install the latest release from GitHub. Use --status to check for updates without installing.",
	RunE: func(cmd *cobra.Command, args []string) error {
		statusOnly, _ := cmd.Flags().GetBool("status")

		fmt.Printf("Current version: %s\n", version.Version)

		release, err := selfupdate.LatestRelease()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !selfupdate.NeedsUpdate(version.Version, release) {
			fmt.Printf("Already up to date (%s)\n", version.Version)
			return nil
		}

		fmt.Printf("New version available: %s\n", release.TagName)

		if statusOnly {
			fmt.Printf("\nRun 'update-watcher self-update' to install the update.\n")
			return nil
		}

		fmt.Printf("Downloading %s...\n", release.TagName)
		if err := selfupdate.DownloadAndReplace(release); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %s\n", err)
			return err
		}

		fmt.Printf("Successfully updated to %s!\n", release.TagName)
		return nil
	},
}

func init() {
	selfUpdateCmd.Flags().Bool("status", false, "only check for updates, do not install")
	rootCmd.AddCommand(selfUpdateCmd)
}
