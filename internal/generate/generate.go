package generate

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/platform"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
)

type Options struct {
	PresetIDs []string
	DryRun    bool
}

func Run(opts Options) error {
	st, err := state.Read(config.StatePath())
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	if len(st.Presets) == 0 {
		ui.Info("no installed presets")
		return nil
	}

	locator := preset.NewLocator()
	proc := preset.NewOutputProcessor()
	manifest, _ := config.ReadManifest()

	scope, err := resolveScope(st, opts.PresetIDs)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return printDryRun(scope, locator, proc, st)
	}

	var totalDeployed int
	for _, id := range scope {
		deployed, err := regeneratePreset(id, locator, proc, manifest)
		if err != nil {
			ui.Warn("generate failed", "preset", id, "error", err)
			continue
		}
		totalDeployed += deployed
	}

	groups, err := preset.CollectConcatOutputs(st, locator, proc)
	if err != nil {
		return fmt.Errorf("collect concat outputs: %w", err)
	}
	if len(groups) > 0 {
		if err := preset.RegenerateConcat(groups); err != nil {
			return fmt.Errorf("regenerate concat: %w", err)
		}
		for _, g := range groups {
			totalDeployed++
			ui.Ok("concat " + g.Target)
		}
	}

	// Generate init scripts
	if err := GenerateInitScripts(); err != nil {
		ui.Warn("init script generation failed", "error", err)
	}

	ui.Ok(fmt.Sprintf("generated %d output(s)", totalDeployed))
	return nil
}

func resolveScope(st *state.State, filterIDs []string) ([]string, error) {
	if len(filterIDs) == 0 {
		var ids []string
		for id, ps := range st.Presets {
			if ps.Status == "installed" {
				ids = append(ids, id)
			}
		}
		return ids, nil
	}

	for _, id := range filterIDs {
		ps, ok := st.Presets[id]
		if !ok {
			return nil, fmt.Errorf("preset %q not found in state", id)
		}
		if ps.Status != "installed" {
			return nil, fmt.Errorf("preset %q status is %q, not installed", id, ps.Status)
		}
	}
	return filterIDs, nil
}

func regeneratePreset(id string, locator *preset.Locator, proc *preset.OutputProcessor, manifest *config.Manifest) (int, error) {
	dir, err := locator.Resolve(id)
	if err != nil {
		return 0, fmt.Errorf("resolve %s: %w", id, err)
	}

	p, err := preset.Load(filepath.Join(dir, "preset.yaml"))
	if err != nil {
		return 0, fmt.Errorf("load %s: %w", id, err)
	}

	cfg := p.Config
	if manifest != nil && manifest.PresetConfig != nil {
		if overrides, ok := manifest.PresetConfig[id]; ok {
			cfg = preset.MergeConfig(p.Config, overrides)
		}
	}

	tplData := buildTemplateData(cfg)
	deployed, err := proc.DeployAll(p.Outputs, dir, tplData)
	if err != nil {
		return 0, err
	}

	if len(deployed) > 0 {
		ui.Ok(fmt.Sprintf("%s: %d output(s)", id, len(deployed)))
	}
	return len(deployed), nil
}

func buildTemplateData(cfg map[string]string) *preset.TemplateData {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, _ := os.UserHomeDir()
		xdg = filepath.Join(home, ".config")
	}

	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}

	return &preset.TemplateData{
		Config:   cfg,
		OS:       platform.DetectOS(),
		Arch:     platform.DetectArch(),
		Home:     os.Getenv("HOME"),
		XDG:      xdg,
		Hostname: hostname,
		User:     username,
		OHP1X:    config.Home(),
	}
}

func printDryRun(scope []string, locator *preset.Locator, proc *preset.OutputProcessor, st *state.State) error {
	ui.Print("[dry-run] Generate plan:")
	ui.Print("")

	for _, id := range scope {
		dir, err := locator.Resolve(id)
		if err != nil {
			continue
		}
		p, err := preset.Load(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			continue
		}

		var lines []string
		for _, o := range p.Outputs {
			mode := preset.InferMode(o)
			target, _ := proc.ResolveTarget(o)
			lines = append(lines, fmt.Sprintf("    %s (%s)", target, mode))
		}
		if len(lines) > 0 {
			ui.Printf("  %s:\n", id)
			for _, l := range lines {
				ui.Print(l)
			}
		}
	}

	groups, err := preset.CollectConcatOutputs(st, locator, proc)
	if err == nil && len(groups) > 0 {
		ui.Print("")
		ui.Print("  Concat targets:")
		for _, g := range groups {
			ui.Printf("    %s (%d sources)\n", g.Target, len(g.Entries))
		}
	}

	ui.Print("")
	ui.Info("dry-run: no changes made")
	return nil
}
