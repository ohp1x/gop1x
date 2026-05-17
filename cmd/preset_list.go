package cmd

import (
	"strings"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available presets with install status",
	Run: func(cmd *cobra.Command, args []string) {
		locator := preset.NewLocator()
		st, err := state.Read(config.StatePath())
		if err != nil {
			ui.Errorf("read state: %v", err)
			return
		}

		presets, err := preset.ListPresets(locator, st)
		if err != nil {
			ui.Errorf("list presets: %v", err)
			return
		}

		if len(presets) == 0 {
			ui.Info("no presets found")
			return
		}

		groups := make(map[string][]preset.PresetInfo)
		for _, p := range presets {
			cat := category(p.Preset.Name)
			groups[cat] = append(groups[cat], p)
		}

		var cats []string
		for cat := range groups {
			cats = append(cats, cat)
		}
		sortStrings(cats)

		for _, cat := range cats {
			ui.Printf("\n%s/\n", cat)
			for _, p := range groups[cat] {
				status := statusIcon(p)
				name := strings.TrimPrefix(p.Preset.Name, cat+"/")
				desc := p.Preset.Description
				if desc == "" {
					desc = "-"
				}
				ui.Printf("  %s %s — %s\n", status, name, desc)
			}
		}
		ui.Print("")
		ui.Printf("Total: %d presets\n", len(presets))
	},
}

func init() {
	presetCmd.AddCommand(presetListCmd)
}

func category(name string) string {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return "other"
}

func statusIcon(p preset.PresetInfo) string {
	switch p.Status {
	case "installed":
		return "\033[32m✓\033[0m"
	case "partial", "failed":
		return "\033[33m!\033[0m"
	default:
		return " "
	}
}

func sortStrings(s []string) {
	for i := range s {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
