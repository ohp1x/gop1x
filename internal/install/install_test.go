package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	dir := filepath.Join(home, "presets", id)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "preset.yaml"), []byte(content), 0644))
}

func TestRun_BasicInstall(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)

	// Write manifest so saveToManifest works
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	opts := Options{
		PresetIDs:  []string{"tools/fzf"},
		DryRun:     false,
		Yes:        true,
		NoSave:     false,
		PkgManager: pm,
	}

	err := Run(opts)
	require.NoError(t, err)

	// Verify state
	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	assert.Equal(t, "user", st.Presets["tools/fzf"].InstalledBy)
	assert.Equal(t, []string{"fzf"}, st.Presets["tools/fzf"].Packages)

	// Verify packages installed
	require.Len(t, pm.InstallCalls, 1)
	assert.Equal(t, []string{"fzf"}, pm.InstallCalls[0])

	// Verify manifest updated
	data, err := os.ReadFile(filepath.Join(home, "ohp1x.yaml"))
	require.NoError(t, err)
	var manifest map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &manifest))
	presets := manifest["presets"].([]interface{})
	assert.Contains(t, presets, "tools/fzf")
}

func TestRun_DryRun(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	opts := Options{
		PresetIDs:  []string{"tools/fzf"},
		DryRun:     true,
		Yes:        true,
		PkgManager: pm,
	}

	err := Run(opts)
	require.NoError(t, err)

	// No state file should be written
	_, err = os.Stat(filepath.Join(home, "local", "gop1x.state.yaml"))
	assert.True(t, os.IsNotExist(err))

	// No packages installed
	assert.Empty(t, pm.InstallCalls)
}

func TestRun_SkipUpToDate(t *testing.T) {
	home := setupTestEnv(t)

	presetContent := `name: fzf
packages:
  - fzf
`
	writePreset(t, home, "tools/fzf", presetContent)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{"fzf": true},
	}

	// First install
	opts := Options{
		PresetIDs:  []string{"tools/fzf"},
		Yes:        true,
		NoSave:     true,
		PkgManager: pm,
	}
	require.NoError(t, Run(opts))

	// Second install — should skip
	pm.InstallCalls = nil
	require.NoError(t, Run(opts))
	assert.Empty(t, pm.InstallCalls)
}

func TestRun_WithDependencies(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "core/system", `name: system
packages:
  - curl
`)
	writePreset(t, home, "shell/zsh", `name: zsh
depends:
  - core/system
packages:
  - zsh
`)
	writePreset(t, home, "tools/fzf", `name: fzf
depends:
  - shell/zsh
packages:
  - fzf
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
	}

	opts := Options{
		PresetIDs:  []string{"tools/fzf"},
		Yes:        true,
		NoSave:     true,
		PkgManager: pm,
	}

	err := Run(opts)
	require.NoError(t, err)

	// All three should be installed
	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["core/system"].Status)
	assert.Equal(t, "installed", st.Presets["shell/zsh"].Status)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)

	// Dependencies marked as "dependency"
	assert.Equal(t, "dependency", st.Presets["core/system"].InstalledBy)
	assert.Equal(t, "dependency", st.Presets["shell/zsh"].InstalledBy)
	assert.Equal(t, "user", st.Presets["tools/fzf"].InstalledBy)
}

func TestRun_ConflictDetection(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "editor/vim", `name: vim
conflicts:
  - editor/emacs
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	// Pre-install emacs in state
	st := state.New()
	st.Presets["editor/emacs"] = state.PresetState{Status: "installed"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true}

	opts := Options{
		PresetIDs:  []string{"editor/vim"},
		Yes:        true,
		PkgManager: pm,
	}

	err := Run(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts")
}

func TestRun_PackageInstallFailure(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/broken", `name: broken
packages:
  - badpkg
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	pm := &pkg.MockManager{
		AvailableTrue: true,
		InstalledMap:  map[string]bool{},
		InstallErr:    assert.AnError,
	}

	opts := Options{
		PresetIDs:  []string{"tools/broken"},
		Yes:        true,
		NoSave:     true,
		PkgManager: pm,
	}

	err := Run(opts)
	require.NoError(t, err) // Run continues past failures

	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "failed", st.Presets["tools/broken"].Status)
}
