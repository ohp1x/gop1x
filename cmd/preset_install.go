package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetInstallCmd = &cobra.Command{
	Use:   "install <preset-id>...",
	Short: "Install preset(s) without saving to manifest",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.Options{
			PresetIDs:  args,
			DryRun:     dryRun,
			Yes:        yes,
			NoSave:     true,
			PkgManager: pm,
		}
		if err := install.Run(opts); err != nil {
			ui.Errorf("install: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	presetCmd.AddCommand(presetInstallCmd)
}
