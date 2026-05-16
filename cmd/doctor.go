package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/doctor"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate environment and report issues",
	Long:  "Check if gop1x environment is properly configured, package managers are available, and presets are valid.",
	Run: func(cmd *cobra.Command, args []string) {
		report := doctor.Run()

		if len(report.Issues) == 0 {
			ui.Ok("All checks passed")
			os.Exit(0)
		}

		for _, issue := range report.Issues {
			switch issue.Level {
			case "error":
				ui.Error(issue.Message)
			case "warning":
				ui.Warn(issue.Message)
			default:
				ui.Info(issue.Message)
			}
		}

		if !report.OK {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
