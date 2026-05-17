package install

import (
	"fmt"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
)

type SyncOptions struct {
	DryRun     bool
	Yes        bool
	PkgManager pkg.PackageManager
}

type SyncPlan struct {
	Install   []string
	Remove    []string
	Reinstall []string
	UpToDate  []string
	Packages  []string
}

type manifest = config.Manifest

func RunSync(opts SyncOptions) error {
	m, err := config.ReadManifest()
	if err != nil {
		return err
	}
	if len(m.Presets) == 0 {
		ui.Info("no presets declared in ohp1x.yaml")
		return nil
	}

	locator := preset.NewLocator()
	st, err := state.Read(config.StatePath())
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	plan, err := computeSyncPlan(m, locator, st, opts.PkgManager)
	if err != nil {
		return err
	}

	if len(plan.Install) == 0 && len(plan.Remove) == 0 && len(plan.Reinstall) == 0 {
		ui.Ok("everything up-to-date")
		return nil
	}

	displaySyncPlan(plan)

	if opts.DryRun {
		ui.Info("dry-run: no changes made")
		return nil
	}

	if !opts.Yes {
		if !ui.Confirm("Proceed?") {
			ui.Info("cancelled")
			return nil
		}
	}

	// Execute removals
	for _, id := range plan.Remove {
		if err := removePreset(id, locator, st); err != nil {
			ui.Error("remove failed", "preset", id, "error", err)
			continue
		}
		ui.Ok("removed " + id)
	}

	// Resolve install order (new + reinstall combined)
	toInstall := append(append([]string{}, plan.Install...), plan.Reinstall...)
	if len(toInstall) > 0 {
		loadFn := func(id string) (*preset.Preset, error) {
			return locator.LoadPreset(id)
		}
		order, err := preset.ResolveOrderMulti(toInstall, loadFn)
		if err != nil {
			return err
		}

		proc := preset.NewOutputProcessor()
		for _, id := range order {
			dir, _ := locator.Resolve(id)
			p, _ := locator.LoadPreset(id)

			mergedConfig := mergeManifestConfigByID(id, p, m.PresetConfig)

			installedBy := "dependency"
			if contains(m.Presets, id) {
				installedBy = "user"
			}

			p.Config = mergedConfig

			if err := installPreset(id, dir, p, opts.PkgManager, st, proc, installedBy); err != nil {
				ui.Error("install failed", "preset", id, "error", err)
				continue
			}
			ui.Ok("installed " + id)
		}
	}

	// Write state
	if err := st.Write(config.StatePath()); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	// Regenerate concat outputs
	proc := preset.NewOutputProcessor()
	groups, err := preset.CollectConcatOutputs(st, locator, proc)
	if err != nil {
		ui.Warn("concat regeneration failed", "error", err)
	} else if len(groups) > 0 {
		if err := preset.RegenerateConcat(groups); err != nil {
			ui.Warn("concat write failed", "error", err)
		}
	}

	return nil
}

func computeSyncPlan(m *manifest, locator *preset.Locator, st *state.State, pm pkg.PackageManager) (*SyncPlan, error) {
	plan := &SyncPlan{}
	desiredSet := make(map[string]bool, len(m.Presets))
	for _, id := range m.Presets {
		desiredSet[id] = true
	}

	// Detect removals: user-installed presets not in manifest
	for id, ps := range st.Presets {
		if ps.InstalledBy != "user" {
			continue
		}
		if !desiredSet[id] {
			plan.Remove = append(plan.Remove, id)
		}
	}

	// Detect installs and reinstalls
	pkgSet := make(map[string]bool)
	for _, id := range m.Presets {
		ps, exists := st.Presets[id]

		if !exists || ps.Status == "failed" || ps.Status == "partial" {
			plan.Install = append(plan.Install, id)
			if err := collectPackages(id, locator, pm, pkgSet, plan); err != nil {
				return nil, err
			}
			continue
		}

		// Check drift
		dir, err := locator.Resolve(id)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", id, err)
		}
		pHash, err := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			return nil, err
		}

		p, err := locator.LoadPreset(id)
		if err != nil {
			return nil, err
		}
		mergedConfig := mergeManifestConfigByID(id, p, m.PresetConfig)
		cHash, _ := preset.ConfigHash(mergedConfig)

		if pHash != ps.PresetHash || cHash != ps.ConfigHash {
			plan.Reinstall = append(plan.Reinstall, id)
			if err := collectPackages(id, locator, pm, pkgSet, plan); err != nil {
				return nil, err
			}
		} else {
			plan.UpToDate = append(plan.UpToDate, id)
		}
	}

	return plan, nil
}

func collectPackages(id string, locator *preset.Locator, pm pkg.PackageManager, pkgSet map[string]bool, plan *SyncPlan) error {
	p, err := locator.LoadPreset(id)
	if err != nil {
		return err
	}
	for _, name := range p.Packages {
		if pkgSet[name] {
			continue
		}
		pkgSet[name] = true
		if pm == nil || !pm.IsInstalled(name) {
			plan.Packages = append(plan.Packages, name)
		}
	}
	return nil
}

func mergeManifestConfigByID(id string, p *preset.Preset, presetConfig map[string]map[string]string) map[string]string {
	if presetConfig == nil {
		return p.Config
	}
	overrides, ok := presetConfig[id]
	if !ok {
		return p.Config
	}
	return preset.MergeConfig(p.Config, overrides)
}

func displaySyncPlan(plan *SyncPlan) {
	ui.Print("Sync plan:")
	if len(plan.Install) > 0 {
		for _, id := range plan.Install {
			ui.Printf("  + %s\n", id)
		}
	}
	if len(plan.Reinstall) > 0 {
		for _, id := range plan.Reinstall {
			ui.Printf("  ~ %s (reinstall)\n", id)
		}
	}
	if len(plan.Remove) > 0 {
		for _, id := range plan.Remove {
			ui.Printf("  - %s\n", id)
		}
	}
	if len(plan.Packages) > 0 {
		ui.Printf("  Packages to install: %v\n", plan.Packages)
	}
	if len(plan.UpToDate) > 0 {
		ui.Printf("  Up-to-date: %v\n", plan.UpToDate)
	}
	ui.Print("")
}

