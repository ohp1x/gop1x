package cmd

import (
	"fmt"
	"os"

	"github.com/ohp1x/gop1x/internal/setup"
	"github.com/spf13/cobra"
)

var (
	initForce    bool
	initSkipSync bool
	initFromURL  string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gop1x environment",
	Long:  "Set up ~/.ohp1x directory structure and initialize git repository",
	Run: func(cmd *cobra.Command, args []string) {
		opts := setup.InitOptions{
			Force:    initForce,
			SkipSync: initSkipSync,
			FromURL:  initFromURL,
		}

		if err := setup.Run(opts); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing initialization")
	initCmd.Flags().BoolVar(&initSkipSync, "skip-sync", false, "skip syncing presets after init")
	initCmd.Flags().StringVar(&initFromURL, "from", "", "clone from existing git repository")
}

