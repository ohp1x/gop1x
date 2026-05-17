package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var upgradeAll bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Re-install drifted presets",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.UpgradeOptions{
			All:        upgradeAll,
			DryRun:     dryRun,
			Yes:        yes,
			PkgManager: pm,
		}
		if err := install.RunUpgrade(opts); err != nil {
			ui.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeAll, "all", false, "Force re-install all presets regardless of drift")
	rootCmd.AddCommand(upgradeCmd)
}
