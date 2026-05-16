package cmd

import (
	"os"

	"github.com/ohp1x/gop1x/internal/install"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var noSave bool

var addCmd = &cobra.Command{
	Use:   "add [preset...]",
	Short: "Install a preset into the environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pm := pkg.NewManager()
		opts := install.Options{
			PresetIDs:  args,
			DryRun:     dryRun,
			Yes:        yes,
			NoSave:     noSave,
			PkgManager: pm,
		}
		if err := install.Run(opts); err != nil {
			ui.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolVar(&noSave, "no-save", false, "install without saving to ohp1x.yaml")
}
