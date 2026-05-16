package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
