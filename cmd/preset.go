package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var presetCmd = &cobra.Command{
	Use:   "preset",
	Short: "preset command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(presetCmd)
}
