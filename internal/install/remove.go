package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/platform"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"gopkg.in/yaml.v3"
)

type RemoveOptions struct {
	PresetIDs  []string
	DryRun     bool
	Yes        bool
	NoSave     bool
	Cascade    bool
	PkgManager pkg.PackageManager
}

type RemovePlan struct {
	Targets        []string
	OrphanDeps     []string
	OrphanPackages []string
}

func RunRemove(opts RemoveOptions) error {
	locator := preset.NewLocator()
	st, err := state.Read(config.StatePath())
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	// Validate all targets are installed
	for _, id := range opts.PresetIDs {
		ps, ok := st.Presets[id]
		if !ok {
			return fmt.Errorf("preset %q is not installed", id)
		}
		// Check required guard
		p, err := locator.LoadPreset(id)
		if err == nil && p.Required && ps.Status == "installed" {
			return fmt.Errorf("preset %q is required, refusing to uninstall", id)
		}
	}

	// Check dependents
	dependents := findDependents(opts.PresetIDs, st, locator)
	if len(dependents) > 0 && !opts.Cascade {
		return fmt.Errorf("cannot remove: presets %v depend on targets; use --cascade to remove them too", dependents)
	}

	// Build full removal list (targets + cascade dependents)
	targets := append([]string{}, opts.PresetIDs...)
	if opts.Cascade {
		for _, dep := range dependents {
			if !contains(targets, dep) {
				targets = append(targets, dep)
			}
		}
	}

	// Detect orphan dependencies
	orphanDeps := findOrphanDeps(targets, st, locator)

	// Detect orphan packages
	orphanPkgs := findOrphanPackages(targets, orphanDeps, st)

	plan := &RemovePlan{
		Targets:        targets,
		OrphanDeps:     orphanDeps,
		OrphanPackages: orphanPkgs,
	}

	displayRemovePlan(plan)

	if opts.DryRun {
		ui.Info("dry-run: no changes made")
		return nil
	}

	if !opts.Yes {
		if ok, err := ui.Confirm("Proceed with removal?"); err != nil {
			return err
		} else if !ok {
			ui.Info("cancelled")
			return nil
		}
	}

	// Prompt for orphan deps removal
	var abortErr error
	removeOrphans := false
	if len(orphanDeps) > 0 && !opts.DryRun {
		if opts.Yes {
			removeOrphans = false
		} else {
			removeOrphans, abortErr = ui.Confirm(fmt.Sprintf("Also remove orphan dependencies %v?", orphanDeps))
			if abortErr != nil {
				return abortErr
			}
		}
	}

	// Prompt for orphan packages removal
	removeOrphanPkgs := false
	if len(orphanPkgs) > 0 && !opts.DryRun {
		if opts.Yes {
			removeOrphanPkgs = false
		} else {
			removeOrphanPkgs, abortErr = ui.Confirm(fmt.Sprintf("Also remove orphan packages %v?", orphanPkgs))
			if abortErr != nil {
				return abortErr
			}
		}
	}

	// Remove each target
	allRemovals := append([]string{}, targets...)
	if removeOrphans {
		allRemovals = append(allRemovals, orphanDeps...)
	}

	for _, id := range allRemovals {
		if err := removePreset(id, locator, st); err != nil {
			ui.Error("remove failed", "preset", id, "error", err)
			continue
		}
		ui.Ok("removed " + id)
	}

	// Remove orphan packages
	if removeOrphanPkgs && opts.PkgManager != nil && len(orphanPkgs) > 0 {
		ui.Info("removing orphan packages", "packages", orphanPkgs)
		if err := opts.PkgManager.Uninstall(orphanPkgs...); err != nil {
			ui.Warn("package uninstall failed", "error", err)
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
	} else {
		if err := preset.RegenerateConcat(groups); err != nil {
			ui.Warn("concat write failed", "error", err)
		}
	}

	// Update manifest
	if !opts.NoSave {
		if err := removeFromManifest(opts.PresetIDs); err != nil {
			ui.Warn("failed to update manifest", "error", err)
		}
	}

	return nil
}

func removePreset(id string, locator *preset.Locator, st *state.State) error {
	ps, ok := st.Presets[id]
	if !ok {
		return nil
	}

	dir, _ := locator.Resolve(id)

	hookEnv := &preset.HookEnv{
		OHP1XHome:    config.Home(),
		PresetID:     id,
		PresetDir:    dir,
		StateDir:     filepath.Dir(config.StatePath()),
		GeneratedDir: config.GeneratedDir(),
		OS:           platform.DetectOS(),
		Arch:         platform.DetectArch(),
	}

	// Load preset for hooks (best-effort — preset dir may be gone)
	p, _ := locator.LoadPreset(id)
	if p != nil {
		hookEnv.Config = p.Config

		// Pre-uninstall hook
		if err := preset.RunHook(p.Hooks.PreUninstall, dir, hookEnv); err != nil {
			ui.Warn("pre_uninstall hook failed", "preset", id, "error", err)
		}
	}

	// Remove deployed outputs
	for _, output := range ps.Outputs {
		if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
			ui.Warn("failed to remove output", "path", output, "error", err)
		}
	}

	// Post-uninstall hook
	if p != nil {
		if err := preset.RunHook(p.Hooks.PostUninstall, dir, hookEnv); err != nil {
			ui.Warn("post_uninstall hook failed", "preset", id, "error", err)
		}
	}

	// Remove from state
	delete(st.Presets, id)
	return nil
}

func findDependents(targets []string, st *state.State, locator *preset.Locator) []string {
	var dependents []string
	for id, ps := range st.Presets {
		if ps.Status != "installed" && ps.Status != "partial" {
			continue
		}
		if contains(targets, id) {
			continue
		}
		p, err := locator.LoadPreset(id)
		if err != nil {
			continue
		}
		for _, dep := range p.Depends {
			if contains(targets, dep) {
				dependents = append(dependents, id)
				break
			}
		}
	}
	return dependents
}

func findOrphanDeps(targets []string, st *state.State, locator *preset.Locator) []string {
	// Collect all deps of targets
	depCandidates := make(map[string]bool)
	for _, id := range targets {
		p, err := locator.LoadPreset(id)
		if err != nil {
			continue
		}
		for _, dep := range p.Depends {
			if !contains(targets, dep) {
				depCandidates[dep] = true
			}
		}
	}

	// Check if any remaining preset still depends on each candidate
	var orphans []string
	for dep := range depCandidates {
		if _, ok := st.Presets[dep]; !ok {
			continue
		}
		stillNeeded := false
		for id, ps := range st.Presets {
			if contains(targets, id) || (ps.Status != "installed" && ps.Status != "partial") {
				continue
			}
			p, err := locator.LoadPreset(id)
			if err != nil {
				continue
			}
			for _, d := range p.Depends {
				if d == dep {
					stillNeeded = true
					break
				}
			}
			if stillNeeded {
				break
			}
		}
		if !stillNeeded {
			orphans = append(orphans, dep)
		}
	}
	return orphans
}

func findOrphanPackages(targets []string, orphanDeps []string, st *state.State) []string {
	// Collect packages from targets + orphan deps
	removedPkgs := make(map[string]bool)
	allRemoved := append(append([]string{}, targets...), orphanDeps...)
	for _, id := range allRemoved {
		if ps, ok := st.Presets[id]; ok {
			for _, p := range ps.Packages {
				removedPkgs[p] = true
			}
		}
	}

	// Check if any remaining preset still uses each package
	for id, ps := range st.Presets {
		if contains(allRemoved, id) {
			continue
		}
		if ps.Status != "installed" && ps.Status != "partial" {
			continue
		}
		for _, p := range ps.Packages {
			delete(removedPkgs, p)
		}
	}

	var orphans []string
	for p := range removedPkgs {
		orphans = append(orphans, p)
	}
	return orphans
}

func displayRemovePlan(plan *RemovePlan) {
	ui.Print("Remove plan:")
	ui.Printf("  Presets to remove: %v\n", plan.Targets)
	if len(plan.OrphanDeps) > 0 {
		ui.Printf("  Orphan dependencies: %v\n", plan.OrphanDeps)
	}
	if len(plan.OrphanPackages) > 0 {
		ui.Printf("  Orphan packages: %v\n", plan.OrphanPackages)
	}
	ui.Print("")
}

func removeFromManifest(presetIDs []string) error {
	manifestPath := filepath.Join(config.Home(), "ohp1x.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return err
	}

	presetsRaw, _ := manifest["presets"]
	var presets []string
	if list, ok := presetsRaw.([]interface{}); ok {
		for _, v := range list {
			if s, ok := v.(string); ok {
				if !contains(presetIDs, s) {
					presets = append(presets, s)
				}
			}
		}
	}
	manifest["presets"] = presets

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, out, 0644)
}
