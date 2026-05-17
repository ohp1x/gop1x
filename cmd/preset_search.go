package cmd

import (
	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search presets by name, tag, or description",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		locator := preset.NewLocator()
		st, err := state.Read(config.StatePath())
		if err != nil {
			ui.Errorf("read state: %v", err)
			return
		}

		results, err := preset.SearchPresets(locator, st, query)
		if err != nil {
			ui.Errorf("search: %v", err)
			return
		}

		if len(results) == 0 {
			ui.Info("no presets matching %q", query)
			return
		}

		for _, p := range results {
			status := statusIcon(p)
			desc := p.Preset.Description
			if desc == "" {
				desc = "-"
			}
			ui.Printf("  %s %s — %s\n", status, p.Preset.Name, desc)
		}
		ui.Printf("\n%d result(s)\n", len(results))
	},
}

func init() {
	presetCmd.AddCommand(presetSearchCmd)
}
