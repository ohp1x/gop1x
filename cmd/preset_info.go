package cmd

import (
	"fmt"
	"strings"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"github.com/spf13/cobra"
)

var presetInfoCmd = &cobra.Command{
	Use:   "info <preset-id>",
	Short: "Show detailed information about a preset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		locator := preset.NewLocator()

		p, err := locator.LoadPreset(id)
		if err != nil {
			ui.Errorf("load preset: %v", err)
			return
		}

		st, err := state.Read(config.StatePath())
		if err != nil {
			ui.Errorf("read state: %v", err)
			return
		}

		ui.Printf("Name:        %s\n", p.Name)
		ui.Printf("Description: %s\n", p.Description)
		if p.Version != "" {
			ui.Printf("Version:     %s\n", p.Version)
		}
		if len(p.OS) > 0 {
			ui.Printf("OS:          %s\n", strings.Join(p.OS, ", "))
		}
		if len(p.Arch) > 0 {
			ui.Printf("Arch:        %s\n", strings.Join(p.Arch, ", "))
		}
		if len(p.Tags) > 0 {
			ui.Printf("Tags:        %s\n", strings.Join(p.Tags, ", "))
		}
		if p.Required {
			ui.Printf("Required:    yes\n")
		}
		if len(p.Depends) > 0 {
			ui.Printf("Depends:     %s\n", strings.Join(p.Depends, ", "))
		}
		if len(p.Recommends) > 0 {
			ui.Printf("Recommends:  %s\n", strings.Join(p.Recommends, ", "))
		}
		if len(p.Conflicts) > 0 {
			ui.Printf("Conflicts:   %s\n", strings.Join(p.Conflicts, ", "))
		}
		if len(p.Packages) > 0 {
			ui.Printf("Packages:    %s\n", strings.Join(p.Packages, ", "))
		}
		if len(p.Outputs) > 0 {
			ui.Printf("Outputs:     %d\n", len(p.Outputs))
		}
		if len(p.Config) > 0 {
			ui.Print("Config:")
			for k, v := range p.Config {
				ui.Printf("  %s: %s\n", k, v)
			}
		}

		ui.Print("")
		if ps, ok := st.Presets[id]; ok {
			ui.Printf("Status:      %s\n", ps.Status)
			ui.Printf("Installed:   %s\n", ps.InstalledAt.Format("2006-01-02 15:04"))
			ui.Printf("Installed by: %s\n", ps.InstalledBy)
		} else {
			fmt.Println("Status:      not installed")
		}
	},
}

func init() {
	presetCmd.AddCommand(presetInfoCmd)
}
