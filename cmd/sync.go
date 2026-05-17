package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Converge installed state to match ohp1x.yaml",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.SyncOptions{
			DryRun:     dryRun,
			Yes:        yes,
			PkgManager: pm,
		}
		if err := install.RunSync(opts); err != nil {
			ui.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
