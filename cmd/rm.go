package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var rmNoSave bool
var rmCascade bool

var rmCmd = &cobra.Command{
	Use:   "rm [preset...]",
	Short: "Remove an installed preset",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.RemoveOptions{
			PresetIDs:  args,
			DryRun:     dryRun,
			Yes:        yes,
			NoSave:     rmNoSave,
			Cascade:    rmCascade,
			PkgManager: pm,
		}
		if err := install.RunRemove(opts); err != nil {
			ui.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVar(&rmNoSave, "no-save", false, "uninstall without removing from ohp1x.yaml")
	rmCmd.Flags().BoolVar(&rmCascade, "cascade", false, "also remove presets that depend on targets")
}
