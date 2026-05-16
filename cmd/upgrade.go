package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
