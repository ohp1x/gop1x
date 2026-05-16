package cmd

import (
	"fmt"
	"os"

	"github.com/ohp1x/gop1x/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate environment and report issues",
	Long:  "Check if gop1x environment is properly configured, package managers are available, and presets are valid.",
	Run: func(cmd *cobra.Command, args []string) {
		report := doctor.Run()

		if len(report.Issues) == 0 {
			fmt.Println("✓ All checks passed")
			os.Exit(0)
		}

		for _, issue := range report.Issues {
			prefix := "ℹ"
			if issue.Level == "warning" {
				prefix = "⚠"
			} else if issue.Level == "error" {
				prefix = "✗"
			}
			fmt.Printf("%s %s\n", prefix, issue.Message)
		}

		if !report.OK {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
