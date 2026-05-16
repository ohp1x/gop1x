package cmd

import (
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Regenerate output files from installed presets",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Warn("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
