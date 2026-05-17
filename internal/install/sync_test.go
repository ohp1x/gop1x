package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeManifest(t *testing.T, home, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte(content), 0644))
}

func TestRunSync_FreshInstall(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writePreset(t, home, "tools/bat", `name: bat
packages:
  - bat
`)
	writeManifest(t, home, `presets:
  - tools/fzf
  - tools/bat
`)

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Equal(t, "installed", st.Presets["tools/bat"].Status)
	assert.Equal(t, "user", st.Presets["tools/fzf"].InstalledBy)
	assert.Equal(t, "user", st.Presets["tools/bat"].InstalledBy)
	assert.Len(t, pm.InstallCalls, 2)
}

func TestRunSync_UpToDate(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	// Pre-seed state as already installed with correct hash
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

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{"fzf": true},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	// No installs should have happened
	assert.Empty(t, pm.InstallCalls)
}

func TestRunSync_RemoveExtra(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writePreset(t, home, "tools/bat", `name: bat
packages:
  - bat
`)

	// Manifest only has fzf
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	// State has both installed
	dir := filepath.Join(home, "presets", "tools", "fzf")
	pHashFzf, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
	cHash, _ := preset.ConfigHash(nil)

	dirBat := filepath.Join(home, "presets", "tools", "bat")
	pHashBat, _ := preset.PresetHash(filepath.Join(dirBat, "preset.yaml"))

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

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{"fzf": true, "bat": true},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Contains(t, st.Presets, "tools/fzf")
	_, hasBat := st.Presets["tools/bat"]
	assert.False(t, hasBat, "tools/bat should be removed")
}

func TestRunSync_ReinstallDrifted(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	// State has old hash (simulating drift)
	cHash, _ := preset.ConfigHash(nil)
	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "deadbeef",
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"fzf"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{"fzf": true},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.NotEqual(t, "deadbeef", st.Presets["tools/fzf"].PresetHash)
}

func TestRunSync_RetryFailed(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	// State has failed status
	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "old",
		ConfigHash:  "old",
		Status:      "failed",
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err = state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Len(t, pm.InstallCalls, 1)
}

func TestRunSync_DryRun(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	err := RunSync(SyncOptions{
		DryRun:     true,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	// State should not exist (no writes)
	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Empty(t, st.Presets)
	assert.Empty(t, pm.InstallCalls)
}

func TestRunSync_WithDependencies(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "core/base", `name: base
packages:
  - coreutils
`)
	writePreset(t, home, "tools/fzf", `name: fzf
depends:
  - core/base
packages:
  - fzf
`)
	// Manifest only lists fzf, but base should be pulled as dependency
	writeManifest(t, home, `presets:
  - tools/fzf
`)

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	err := RunSync(SyncOptions{
		DryRun:     false,
		Yes:        true,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Equal(t, "installed", st.Presets["core/base"].Status)
	assert.Equal(t, "user", st.Presets["tools/fzf"].InstalledBy)
	assert.Equal(t, "dependency", st.Presets["core/base"].InstalledBy)
}
