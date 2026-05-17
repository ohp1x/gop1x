package install

import (
	"path/filepath"
	"testing"

	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunUpgrade_NoPresets(t *testing.T) {
	setupTestEnv(t)

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{}}
	err := RunUpgrade(UpgradeOptions{
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)
	assert.Empty(t, pm.InstallCalls)
}

func TestRunUpgrade_AllUpToDate(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n")

	dir := filepath.Join(home, "presets", "tools", "fzf")
	pHash, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
	cHash, _ := preset.ConfigHash(nil)

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  pHash,
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"fzf"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{"fzf": true}}
	err := RunUpgrade(UpgradeOptions{
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)
	assert.Empty(t, pm.InstallCalls)
}

func TestRunUpgrade_DriftedPreset(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n")

	cHash, _ := preset.ConfigHash(nil)
	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "oldhash1",
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"fzf"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{"fzf": true}}
	err := RunUpgrade(UpgradeOptions{
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.NotEqual(t, "oldhash1", st.Presets["tools/fzf"].PresetHash)
}

func TestRunUpgrade_ConfigDrift(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "lang/go", `name: go
config:
  version: "1.22"
packages:
  - golang
`)
	writeManifest(t, home, `presets:
  - lang/go
preset_config:
  lang/go:
    version: "1.24"
`)

	dir := filepath.Join(home, "presets", "lang", "go")
	pHash, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
	oldConfig := map[string]string{"version": "1.22"}
	cHash, _ := preset.ConfigHash(oldConfig)

	st := state.New()
	st.Presets["lang/go"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  pHash,
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"golang"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{"golang": true}}
	err := RunUpgrade(UpgradeOptions{
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["lang/go"].Status)
	// Config hash should now reflect merged config with version=1.24
	newConfig := map[string]string{"version": "1.24"}
	expectedHash, _ := preset.ConfigHash(newConfig)
	assert.Equal(t, expectedHash, st.Presets["lang/go"].ConfigHash)
}

func TestRunUpgrade_RetryFailed(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n")

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "old",
		ConfigHash:  "old",
		Status:      "failed",
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{}}
	err := RunUpgrade(UpgradeOptions{
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Len(t, pm.InstallCalls, 1)
}

func TestRunUpgrade_All(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writePreset(t, home, "tools/bat", `name: bat
packages:
  - bat
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n  - tools/bat\n")

	dir := filepath.Join(home, "presets", "tools", "fzf")
	pHashFzf, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
	dirBat := filepath.Join(home, "presets", "tools", "bat")
	pHashBat, _ := preset.PresetHash(filepath.Join(dirBat, "preset.yaml"))
	cHash, _ := preset.ConfigHash(nil)

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  pHashFzf,
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"fzf"},
	}
	st.Presets["tools/bat"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  pHashBat,
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"bat"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{"fzf": true, "bat": true}}
	err := RunUpgrade(UpgradeOptions{
		All:        true,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Equal(t, "installed", st.Presets["tools/bat"].Status)
}

func TestRunUpgrade_DryRun(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n")

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "oldhash1",
		ConfigHash:  "oldcfg1",
		Status:      "installed",
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{}}
	err := RunUpgrade(UpgradeOptions{
		DryRun:     true,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	// State should remain unchanged
	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "oldhash1", st.Presets["tools/fzf"].PresetHash)
	assert.Empty(t, pm.InstallCalls)
}
