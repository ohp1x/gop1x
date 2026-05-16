package cmd

import (
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [preset]",
	Short: "Install a preset into the environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.Warn("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
