package status

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

func setupTestEnv(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("OHP1X_HOME", home)

	dirs := []string{
		filepath.Join(home, "presets"),
		filepath.Join(home, "local", "presets"),
		filepath.Join(home, "local"),
		filepath.Join(home, "generated"),
	}
	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, 0755))
	}
	return home
}

func writePreset(t *testing.T, home, id, content string) {
	t.Helper()
	dir := filepath.Join(home, "presets", id)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "preset.yaml"), []byte(content), 0644))
}

func writeManifest(t *testing.T, home, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte(content), 0644))
}

func TestCollect_EmptyState(t *testing.T) {
	setupTestEnv(t)

	pm := &pkg.MockManager{AvailableTrue: true}
	r, err := Collect(Options{PkgManager: pm})
	require.NoError(t, err)

	assert.Empty(t, r.Presets)
	assert.Empty(t, r.Drifted)
	assert.Empty(t, r.Failed)
	assert.Empty(t, r.Partial)
}

func TestCollect_UpToDate(t *testing.T) {
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

	pm := &pkg.MockManager{AvailableTrue: true}
	r, err := Collect(Options{PkgManager: pm})
	require.NoError(t, err)

	assert.Len(t, r.Presets, 1)
	assert.Equal(t, "tools/fzf", r.Presets[0].ID)
	assert.Equal(t, "installed", r.Presets[0].Status)
	assert.Empty(t, r.Drifted)
}

func TestCollect_DriftedPreset(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writeManifest(t, home, "presets:\n  - tools/fzf\n")

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  "deadbeef",
		ConfigHash:  "00000000",
		Status:      "installed",
		Packages:    []string{"fzf"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true}
	r, err := Collect(Options{PkgManager: pm})
	require.NoError(t, err)

	assert.Len(t, r.Drifted, 1)
	assert.Equal(t, "tools/fzf", r.Drifted[0].ID)
	assert.Equal(t, "preset changed", r.Drifted[0].Reason)
}

func TestCollect_FailedAndPartial(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	writePreset(t, home, "tools/bat", `name: bat
packages:
  - bat
`)

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		InstalledBy: "user",
		Status:      "failed",
	}
	st.Presets["tools/bat"] = state.PresetState{
		InstalledBy: "user",
		Status:      "partial",
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true}
	r, err := Collect(Options{PkgManager: pm})
	require.NoError(t, err)

	assert.Len(t, r.Failed, 1)
	assert.Contains(t, r.Failed, "tools/fzf")
	assert.Len(t, r.Partial, 1)
	assert.Contains(t, r.Partial, "tools/bat")
	assert.Empty(t, r.Drifted)
}

func TestCollect_ConfigDrift(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "lang/go", `name: go
config:
  version: "1.22"
packages:
  - go
`)
	writeManifest(t, home, `presets:
  - lang/go
preset_config:
  lang/go:
    version: "1.24"
`)

	dir := filepath.Join(home, "presets", "lang", "go")
	pHash, _ := preset.PresetHash(filepath.Join(dir, "preset.yaml"))
	// Config hash was computed with old config (version=1.22, no override)
	oldConfig := map[string]string{"version": "1.22"}
	cHash, _ := preset.ConfigHash(oldConfig)

	st := state.New()
	st.Presets["lang/go"] = state.PresetState{
		InstalledBy: "user",
		PresetHash:  pHash,
		ConfigHash:  cHash,
		Status:      "installed",
		Packages:    []string{"go"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true}
	r, err := Collect(Options{PkgManager: pm})
	require.NoError(t, err)

	assert.Len(t, r.Drifted, 1)
	assert.Equal(t, "lang/go", r.Drifted[0].ID)
	assert.Equal(t, "config changed", r.Drifted[0].Reason)
}
