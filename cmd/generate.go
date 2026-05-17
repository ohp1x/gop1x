package cmd

import (
	"github.com/ohp1x/gop1x/internal/generate"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var generatePresets []string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Regenerate output files from installed presets",
	Long:  "Regenerate all outputs (concat, copy, template, symlink) from installed presets. Use --preset to limit scope.",
	Run: func(cmd *cobra.Command, args []string) {
		opts := generate.Options{
			PresetIDs: generatePresets,
			DryRun:    dryRun,
		}
		if err := generate.Run(opts); err != nil {
			ui.Error("generate failed", "error", err)
		}
	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&generatePresets, "preset", nil, "regenerate only specific preset(s)")
	rootCmd.AddCommand(generateCmd)
}
