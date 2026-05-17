package status

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/platform"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"gopkg.in/yaml.v3"
)

type Options struct {
	PkgManager pkg.PackageManager
}

type DriftEntry struct {
	ID     string
	Reason string
}

type Report struct {
	Home       string
	OS         string
	Arch       string
	PkgMgr    string
	Presets    []PresetEntry
	Drifted    []DriftEntry
	Failed     []string
	Partial    []string
}

type PresetEntry struct {
	ID          string
	Status      string
	InstalledBy string
}

func Collect(opts Options) (*Report, error) {
	r := &Report{
		Home: config.Home(),
		OS:   platform.DetectOS(),
		Arch: platform.DetectArch(),
	}
	if opts.PkgManager != nil {
		r.PkgMgr = opts.PkgManager.Name()
	}

	st, err := state.Read(config.StatePath())
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	if len(st.Presets) == 0 {
		return r, nil
	}

	ids := make([]string, 0, len(st.Presets))
	for id := range st.Presets {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	locator := preset.NewLocator()
	manifestConfig := readPresetConfig()

	for _, id := range ids {
		ps := st.Presets[id]
		r.Presets = append(r.Presets, PresetEntry{
			ID:          id,
			Status:      ps.Status,
			InstalledBy: ps.InstalledBy,
		})

		switch ps.Status {
		case "failed":
			r.Failed = append(r.Failed, id)
			continue
		case "partial":
			r.Partial = append(r.Partial, id)
			continue
		}

		if ps.Status != "installed" {
			continue
		}

		dir, err := locator.Resolve(id)
		if err != nil {
			continue
		}

		pHash, err := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			continue
		}

		if pHash != ps.PresetHash {
			r.Drifted = append(r.Drifted, DriftEntry{ID: id, Reason: "preset changed"})
			continue
		}

		p, err := locator.LoadPreset(id)
		if err != nil {
			continue
		}
		merged := mergeWithManifest(id, p.Config, manifestConfig)
		cHash, _ := preset.ConfigHash(merged)
		if cHash != ps.ConfigHash {
			r.Drifted = append(r.Drifted, DriftEntry{ID: id, Reason: "config changed"})
		}
	}

	return r, nil
}

func Display(r *Report) {
	ui.Print("System:")
	ui.Printf("  Home:     %s\n", r.Home)
	ui.Printf("  OS:       %s/%s\n", r.OS, r.Arch)
	if r.PkgMgr != "" {
		ui.Printf("  Packages: %s\n", r.PkgMgr)
	}
	ui.Print("")

	if len(r.Presets) == 0 {
		ui.Print("No presets installed.")
		return
	}

	ui.Printf("Presets (%d installed):\n", len(r.Presets))
	for _, p := range r.Presets {
		ui.Printf("  %-20s %-12s %s\n", p.ID, p.Status, p.InstalledBy)
	}
	ui.Print("")

	if len(r.Drifted) > 0 {
		ui.Printf("Drift (%d drifted):\n", len(r.Drifted))
		for _, d := range r.Drifted {
			ui.Printf("  ~ %-20s %s\n", d.ID, d.Reason)
		}
		ui.Print("")
	}

	if len(r.Failed) > 0 {
		ui.Printf("Failed: %v\n", r.Failed)
	}
	if len(r.Partial) > 0 {
		ui.Printf("Partial: %v\n", r.Partial)
	}

	if len(r.Drifted) == 0 && len(r.Failed) == 0 && len(r.Partial) == 0 {
		ui.Ok("all presets up-to-date")
	} else {
		parts := []string{}
		if len(r.Drifted) > 0 {
			parts = append(parts, fmt.Sprintf("%d drifted", len(r.Drifted)))
		}
		if len(r.Failed)+len(r.Partial) > 0 {
			parts = append(parts, fmt.Sprintf("%d failed/partial", len(r.Failed)+len(r.Partial)))
		}
		ui.Warn("run 'gop1x sync' to converge", "issues", fmt.Sprintf("%v", parts))
	}
}

func Run(opts Options) error {
	r, err := Collect(opts)
	if err != nil {
		return err
	}
	Display(r)
	return nil
}

func readPresetConfig() map[string]map[string]string {
	manifestPath := filepath.Join(config.Home(), "ohp1x.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}
	var m struct {
		PresetConfig map[string]map[string]string `yaml:"preset_config"`
	}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m.PresetConfig
}

func mergeWithManifest(id string, defaults map[string]string, manifestConfig map[string]map[string]string) map[string]string {
	if manifestConfig == nil {
		return defaults
	}
	overrides, ok := manifestConfig[id]
	if !ok {
		return defaults
	}
	return preset.MergeConfig(defaults, overrides)
}
