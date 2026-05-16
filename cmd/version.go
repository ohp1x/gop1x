package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ohp1x/gop1x/internal/ui"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Printf("gop1x %s (commit: %s, built: %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
