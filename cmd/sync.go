package cmd

import (
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Converge installed state to match ohp1x.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Warn("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
