package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/platform"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
	"gopkg.in/yaml.v3"
)

type Options struct {
	PresetIDs  []string
	DryRun     bool
	Yes        bool
	NoSave     bool
	PkgManager pkg.PackageManager
}

type Plan struct {
	Order    []string
	Packages []string
	Skipped  []string
	Conflicts []string
}

func Run(opts Options) error {
	locator := preset.NewLocator()
	st, err := state.Read(config.StatePath())
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	// Validate OS/arch compatibility for root presets
	for _, id := range opts.PresetIDs {
		p, err := locator.LoadPreset(id)
		if err != nil {
			return err
		}
		if err := checkCompat(p); err != nil {
			return fmt.Errorf("preset %s: %w", id, err)
		}
	}

	// Resolve DAG
	loadFn := func(id string) (*preset.Preset, error) {
		return locator.LoadPreset(id)
	}
	order, err := preset.ResolveOrderMulti(opts.PresetIDs, loadFn)
	if err != nil {
		return err
	}

	// Compute plan
	plan, err := computePlan(order, locator, st, opts.PkgManager)
	if err != nil {
		return err
	}

	// Check conflicts
	if err := checkConflicts(order, locator, st); err != nil {
		return err
	}

	// Display plan
	displayPlan(plan, opts.PresetIDs)

	// Collect and prompt recommends
	recommends := collectRecommends(opts.PresetIDs, plan.Order, locator, st)
	selected := promptRecommends(recommends, opts.Yes || opts.DryRun)
	if len(selected) > 0 {
		expanded := append(append([]string{}, opts.PresetIDs...), selected...)
		order, err = preset.ResolveOrderMulti(expanded, loadFn)
		if err != nil {
			return err
		}
		plan, err = computePlan(order, locator, st, opts.PkgManager)
		if err != nil {
			return err
		}
		if err := checkConflicts(order, locator, st); err != nil {
			return err
		}
		displayPlan(plan, expanded)
	}

	if opts.DryRun {
		ui.Info("dry-run: no changes made")
		return nil
	}

	// Confirm
	if !opts.Yes {
		if !confirm("Proceed with installation?") {
			ui.Info("cancelled")
			return nil
		}
	}

	// Install each preset in topological order
	proc := preset.NewOutputProcessor()
	allUserIDs := append(append([]string{}, opts.PresetIDs...), selected...)

	manifest, _ := config.ReadManifest()

	for _, id := range plan.Order {
		if contains(plan.Skipped, id) {
			ui.Info("skip (up-to-date)", "preset", id)
			continue
		}

		dir, _ := locator.Resolve(id)
		p, _ := locator.LoadPreset(id)

		if manifest != nil && manifest.PresetConfig != nil {
			if overrides, ok := manifest.PresetConfig[id]; ok {
				p.Config = preset.MergeConfig(p.Config, overrides)
			}
		}

		installedBy := "dependency"
		if contains(allUserIDs, id) {
			installedBy = "user"
		}

		if err := installPreset(id, dir, p, opts.PkgManager, st, proc, installedBy); err != nil {
			ui.Error("install failed", "preset", id, "error", err)
			continue
		}

		ui.Ok("installed " + id)
	}

	// Write state
	if err := st.Write(config.StatePath()); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	// Regenerate concat outputs
	groups, err := preset.CollectConcatOutputs(st, locator, proc)
	if err != nil {
		ui.Warn("concat regeneration failed", "error", err)
	} else if len(groups) > 0 {
		if err := preset.RegenerateConcat(groups); err != nil {
			ui.Warn("concat write failed", "error", err)
		}
	}

	// Save to manifest
	if !opts.NoSave {
		if err := saveToManifest(allUserIDs); err != nil {
			ui.Warn("failed to save to manifest", "error", err)
		}
	}

	return nil
}

func computePlan(order []string, locator *preset.Locator, st *state.State, pm pkg.PackageManager) (*Plan, error) {
	plan := &Plan{Order: order}
	pkgSet := make(map[string]bool)

	for _, id := range order {
		dir, err := locator.Resolve(id)
		if err != nil {
			return nil, err
		}

		p, err := preset.Load(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			return nil, err
		}

		// Hash check — skip if already installed with same hash
		if ps, ok := st.Presets[id]; ok && ps.Status == "installed" {
			h, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
			if h == ps.PresetHash {
				plan.Skipped = append(plan.Skipped, id)
				continue
			}
		}

		for _, pkg := range p.Packages {
			if !pkgSet[pkg] {
				pkgSet[pkg] = true
				if pm == nil || !pm.IsInstalled(pkg) {
					plan.Packages = append(plan.Packages, pkg)
				}
			}
		}
	}

	return plan, nil
}

func checkCompat(p *preset.Preset) error {
	if len(p.OS) > 0 {
		currentOS := platform.DetectOS()
		if !contains(p.OS, currentOS) {
			return fmt.Errorf("incompatible OS: requires %v, got %s", p.OS, currentOS)
		}
	}
	if len(p.Arch) > 0 {
		currentArch := platform.DetectArch()
		if !contains(p.Arch, currentArch) {
			return fmt.Errorf("incompatible arch: requires %v, got %s", p.Arch, currentArch)
		}
	}
	return nil
}

func checkConflicts(order []string, locator *preset.Locator, st *state.State) error {
	for _, id := range order {
		p, err := locator.LoadPreset(id)
		if err != nil {
			continue
		}
		for _, conflict := range p.Conflicts {
			if ps, ok := st.Presets[conflict]; ok && ps.Status == "installed" {
				return fmt.Errorf("preset %s conflicts with installed preset %s", id, conflict)
			}
		}
	}
	return nil
}

func displayPlan(plan *Plan, rootIDs []string) {
	ui.Print("Install plan:")
	ui.Printf("  Presets: %s\n", strings.Join(plan.Order, ", "))
	if len(plan.Skipped) > 0 {
		ui.Printf("  Skipped (up-to-date): %s\n", strings.Join(plan.Skipped, ", "))
	}
	if len(plan.Packages) > 0 {
		ui.Printf("  Packages to install: %s\n", strings.Join(plan.Packages, ", "))
	} else {
		ui.Print("  Packages: (none)")
	}
	ui.Print("")
}

func installPreset(id, dir string, p *preset.Preset, pm pkg.PackageManager, st *state.State, proc *preset.OutputProcessor, installedBy string) error {
	presetYAML := filepath.Join(dir, "preset.yaml")
	pHash, _ := preset.PresetHash(presetYAML)
	cHash, _ := preset.ConfigHash(p.Config)

	hookEnv := &preset.HookEnv{
		OHP1XHome:    config.Home(),
		PresetID:     id,
		PresetDir:    dir,
		StateDir:     filepath.Dir(config.StatePath()),
		GeneratedDir: config.GeneratedDir(),
		OS:           platform.DetectOS(),
		Arch:         platform.DetectArch(),
		Config:       p.Config,
	}
	if pm != nil {
		hookEnv.PkgManager = pm.Name()
	}

	// Pre-install hook
	if err := preset.RunHook(p.Hooks.PreInstall, dir, hookEnv); err != nil {
		st.Presets[id] = state.PresetState{
			InstalledAt: time.Now(),
			InstalledBy: installedBy,
			PresetHash:  pHash,
			ConfigHash:  cHash,
			Status:      "failed",
		}
		return fmt.Errorf("pre_install hook: %w", err)
	}

	// Install packages
	if len(p.Packages) > 0 && pm != nil {
		var toInstall []string
		for _, pkg := range p.Packages {
			if !pm.IsInstalled(pkg) {
				toInstall = append(toInstall, pkg)
			}
		}
		if len(toInstall) > 0 {
			if err := pm.Install(toInstall...); err != nil {
				st.Presets[id] = state.PresetState{
					InstalledAt: time.Now(),
					InstalledBy: installedBy,
					PresetHash:  pHash,
					ConfigHash:  cHash,
					Packages:    toInstall,
					Status:      "failed",
				}
				return fmt.Errorf("package install: %w", err)
			}
		}
	}

	// Brewfile
	brewfilePath := filepath.Join(dir, "Brewfile")
	if _, err := os.Stat(brewfilePath); err == nil {
		if pm != nil && pm.Name() == "brew" {
			cmd := fmt.Sprintf("brew bundle --file=%s --no-lock", brewfilePath)
			ui.Info("running " + cmd)
			// Brewfile execution delegated to shell
			if err := runShell(cmd, dir); err != nil {
				ui.Warn("Brewfile failed", "error", err)
			}
		}
	}

	// Deploy outputs
	tplData := &preset.TemplateData{
		Config:   p.Config,
		OS:       platform.DetectOS(),
		Arch:     platform.DetectArch(),
		Home:     os.Getenv("HOME"),
		XDG:      os.Getenv("XDG_CONFIG_HOME"),
		OHP1X:    config.Home(),
	}
	deployed, err := proc.DeployAll(p.Outputs, dir, tplData)
	if err != nil {
		st.Presets[id] = state.PresetState{
			InstalledAt: time.Now(),
			InstalledBy: installedBy,
			PresetHash:  pHash,
			ConfigHash:  cHash,
			Packages:    p.Packages,
			Outputs:     deployed,
			Status:      "partial",
		}
		return fmt.Errorf("deploy outputs: %w", err)
	}

	// Post-install hook
	if err := preset.RunHook(p.Hooks.PostInstall, dir, hookEnv); err != nil {
		st.Presets[id] = state.PresetState{
			InstalledAt: time.Now(),
			InstalledBy: installedBy,
			PresetHash:  pHash,
			ConfigHash:  cHash,
			Packages:    p.Packages,
			Outputs:     deployed,
			Status:      "partial",
		}
		return fmt.Errorf("post_install hook: %w", err)
	}

	// Success
	st.Presets[id] = state.PresetState{
		InstalledAt: time.Now(),
		InstalledBy: installedBy,
		PresetHash:  pHash,
		ConfigHash:  cHash,
		Packages:    p.Packages,
		Outputs:     deployed,
		Status:      "installed",
	}
	return nil
}

func saveToManifest(presetIDs []string) error {
	manifestPath := filepath.Join(config.Home(), "ohp1x.yaml")

	var manifest map[string]interface{}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		manifest = make(map[string]interface{})
	} else {
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			return err
		}
	}

	presetsRaw, _ := manifest["presets"]
	var presets []string
	if list, ok := presetsRaw.([]interface{}); ok {
		for _, v := range list {
			if s, ok := v.(string); ok {
				presets = append(presets, s)
			}
		}
	}

	for _, id := range presetIDs {
		if !contains(presets, id) {
			presets = append(presets, id)
		}
	}
	manifest["presets"] = presets

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, out, 0644)
}

func runShell(command, dir string) error {
	cmd := newCommand("/bin/sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func confirm(prompt string) bool {
	ui.Printf("%s [y/N] ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
