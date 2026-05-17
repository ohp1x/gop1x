package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetUninstallCascade bool

var presetUninstallCmd = &cobra.Command{
	Use:   "uninstall <preset-id>...",
	Short: "Uninstall preset(s) without modifying manifest",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.RemoveOptions{
			PresetIDs:  args,
			DryRun:     dryRun,
			Yes:        yes,
			NoSave:     true,
			Cascade:    presetUninstallCascade,
			PkgManager: pm,
		}
		if err := install.RunRemove(opts); err != nil {
			ui.Errorf("uninstall: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	presetUninstallCmd.Flags().BoolVar(&presetUninstallCascade, "cascade", false, "also uninstall dependents")
	presetCmd.AddCommand(presetUninstallCmd)
}
