package cmd

import (
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed presets and detect drift",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Warn("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
