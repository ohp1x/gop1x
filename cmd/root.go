package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ohp1x/gop1x/internal/ui"
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
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		os.Exit(130)
	}()

	if err := rootCmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes without executing")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompts")
}
