package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
