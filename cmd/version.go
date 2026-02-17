package cmd

import (
	"fmt"

	"github.com/mahype/update-watcher/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(version.Version)
			return
		}
		fmt.Printf("update-watcher %s (commit: %s, built: %s)\n", version.Version, version.Commit, version.Date)
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "print only the version number")
	rootCmd.AddCommand(versionCmd)
}
