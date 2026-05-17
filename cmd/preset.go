package cmd

import (
	"github.com/spf13/cobra"
)

var presetCmd = &cobra.Command{
	Use:   "preset [command]",
	Short: "Manage presets (list, info, search, install, uninstall)",
}

func init() {
	rootCmd.AddCommand(presetCmd)
}
