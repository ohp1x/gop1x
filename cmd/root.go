package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"

	dryRun bool
	yes    bool
)

var rootCmd = &cobra.Command{
	Use:   "gop1x",
	Short: "Developer workstation provisioner & preset toolkit",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes without executing")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompts")
}
