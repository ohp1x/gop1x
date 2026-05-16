package cmd

import (
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetCmd = &cobra.Command{
	Use:   "preset [command]",
	Short: "Manage presets (list, info, search)",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Warn("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(presetCmd)
}
