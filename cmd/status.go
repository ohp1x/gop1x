package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/status"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed presets and detect drift",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := status.Options{
			PkgManager: pm,
		}
		if err := status.Run(opts); err != nil {
			ui.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
