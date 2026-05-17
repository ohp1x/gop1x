package install

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
)

type UpgradeOptions struct {
	All        bool
	DryRun     bool
	Yes        bool
	PkgManager pkg.PackageManager
}

type UpgradePlan struct {
	Reinstall []string
	UpToDate  []string
	Packages  []string
}

func RunUpgrade(opts UpgradeOptions) error {
	st, err := state.Read(config.StatePath())
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	if len(st.Presets) == 0 {
		ui.Info("nothing to upgrade — no presets installed")
		return nil
	}

	locator := preset.NewLocator()
	m, _ := config.ReadManifest()
	var presetConfig map[string]map[string]string
	if m != nil {
		presetConfig = m.PresetConfig
	}

	plan, err := computeUpgradePlan(st, locator, presetConfig, opts)
	if err != nil {
		return err
	}

	if len(plan.Reinstall) == 0 {
		ui.Ok("all presets up-to-date")
		return nil
	}

	displayUpgradePlan(plan)

	if opts.DryRun {
		ui.Info("dry-run: no changes made")
		return nil
	}

	if !opts.Yes {
		if !ui.Confirm("Proceed with upgrade?") {
			ui.Info("cancelled")
			return nil
		}
	}

	loadFn := func(id string) (*preset.Preset, error) {
		return locator.LoadPreset(id)
	}
	order, err := preset.ResolveOrderMulti(plan.Reinstall, loadFn)
	if err != nil {
		return err
	}

	proc := preset.NewOutputProcessor()
	for _, id := range order {
		dir, _ := locator.Resolve(id)
		p, _ := locator.LoadPreset(id)

		mergedConfig := mergeManifestConfigByID(id, p, presetConfig)
		p.Config = mergedConfig

		installedBy := "dependency"
		if ps, ok := st.Presets[id]; ok {
			installedBy = ps.InstalledBy
		}

		if err := installPreset(id, dir, p, opts.PkgManager, st, proc, installedBy); err != nil {
			ui.Error("upgrade failed", "preset", id, "error", err)
			continue
		}
		ui.Ok("upgraded " + id)
	}

	if err := st.Write(config.StatePath()); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

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

func computeUpgradePlan(st *state.State, locator *preset.Locator, presetConfig map[string]map[string]string, opts UpgradeOptions) (*UpgradePlan, error) {
	plan := &UpgradePlan{}
	pkgSet := make(map[string]bool)

	ids := make([]string, 0, len(st.Presets))
	for id := range st.Presets {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		ps := st.Presets[id]

		if opts.All {
			plan.Reinstall = append(plan.Reinstall, id)
			collectUpgradePackages(id, locator, opts.PkgManager, pkgSet, plan)
			continue
		}

		if ps.Status == "failed" || ps.Status == "partial" {
			plan.Reinstall = append(plan.Reinstall, id)
			collectUpgradePackages(id, locator, opts.PkgManager, pkgSet, plan)
			continue
		}

		if ps.Status != "installed" {
			continue
		}

		dir, err := locator.Resolve(id)
		if err != nil {
			plan.UpToDate = append(plan.UpToDate, id)
			continue
		}

		pHash, err := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			plan.UpToDate = append(plan.UpToDate, id)
			continue
		}

		if pHash != ps.PresetHash {
			plan.Reinstall = append(plan.Reinstall, id)
			collectUpgradePackages(id, locator, opts.PkgManager, pkgSet, plan)
			continue
		}

		p, err := locator.LoadPreset(id)
		if err != nil {
			plan.UpToDate = append(plan.UpToDate, id)
			continue
		}
		merged := mergeManifestConfigByID(id, p, presetConfig)
		cHash, _ := preset.ConfigHash(merged)
		if cHash != ps.ConfigHash {
			plan.Reinstall = append(plan.Reinstall, id)
			collectUpgradePackages(id, locator, opts.PkgManager, pkgSet, plan)
			continue
		}

		plan.UpToDate = append(plan.UpToDate, id)
	}

	return plan, nil
}

func collectUpgradePackages(id string, locator *preset.Locator, pm pkg.PackageManager, pkgSet map[string]bool, plan *UpgradePlan) {
	p, err := locator.LoadPreset(id)
	if err != nil {
		return
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
}

func displayUpgradePlan(plan *UpgradePlan) {
	ui.Print("Upgrade plan:")
	for _, id := range plan.Reinstall {
		ui.Printf("  ~ %s\n", id)
	}
	if len(plan.Packages) > 0 {
		ui.Printf("  Packages to install: %v\n", plan.Packages)
	}
	if len(plan.UpToDate) > 0 {
		ui.Printf("  Up-to-date (skip): %v\n", plan.UpToDate)
	}
	ui.Print("")
}

