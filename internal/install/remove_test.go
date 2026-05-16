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

func TestRunRemove_Basic(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets:\n  - tools/fzf\n"), 0644))

	// Pre-install
	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		Status:     "installed",
		InstalledBy: "user",
		Packages:   []string{"fzf"},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	pm := &pkg.MockManager{AvailableTrue: true}
	opts := RemoveOptions{
		PresetIDs:  []string{"tools/fzf"},
		Yes:        true,
		PkgManager: pm,
	}

	err := RunRemove(opts)
	require.NoError(t, err)

	// Verify state cleared
	st2, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	_, exists := st2.Presets["tools/fzf"]
	assert.False(t, exists)

	// Verify manifest updated
	data, err := os.ReadFile(filepath.Join(home, "ohp1x.yaml"))
	require.NoError(t, err)
	var manifest map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &manifest))
	presets, _ := manifest["presets"].([]interface{})
	assert.Empty(t, presets)
}

func TestRunRemove_NotInstalled(t *testing.T) {
	home := setupTestEnv(t)
	writePreset(t, home, "tools/fzf", `name: fzf`)

	opts := RemoveOptions{
		PresetIDs: []string{"tools/fzf"},
		Yes:       true,
	}

	err := RunRemove(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestRunRemove_RequiredGuard(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "core/system", `name: system
required: true
`)

	st := state.New()
	st.Presets["core/system"] = state.PresetState{Status: "installed"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	opts := RemoveOptions{
		PresetIDs: []string{"core/system"},
		Yes:       true,
	}

	err := RunRemove(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestRunRemove_DependentBlocks(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "shell/zsh", `name: zsh`)
	writePreset(t, home, "tools/fzf", `name: fzf
depends:
  - shell/zsh
`)

	st := state.New()
	st.Presets["shell/zsh"] = state.PresetState{Status: "installed"}
	st.Presets["tools/fzf"] = state.PresetState{Status: "installed"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	opts := RemoveOptions{
		PresetIDs: []string{"shell/zsh"},
		Yes:       true,
	}

	err := RunRemove(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "depend")
}

func TestRunRemove_Cascade(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "shell/zsh", `name: zsh`)
	writePreset(t, home, "tools/fzf", `name: fzf
depends:
  - shell/zsh
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets:\n  - shell/zsh\n  - tools/fzf\n"), 0644))

	st := state.New()
	st.Presets["shell/zsh"] = state.PresetState{Status: "installed", InstalledBy: "user"}
	st.Presets["tools/fzf"] = state.PresetState{Status: "installed", InstalledBy: "user"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	opts := RemoveOptions{
		PresetIDs: []string{"shell/zsh"},
		Yes:       true,
		Cascade:   true,
	}

	err := RunRemove(opts)
	require.NoError(t, err)

	st2, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Empty(t, st2.Presets)
}

func TestRunRemove_DryRun(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf`)

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{Status: "installed"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	opts := RemoveOptions{
		PresetIDs: []string{"tools/fzf"},
		DryRun:    true,
		Yes:       true,
	}

	err := RunRemove(opts)
	require.NoError(t, err)

	// State should still have the preset
	st2, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	_, exists := st2.Presets["tools/fzf"]
	assert.True(t, exists)
}

func TestRunRemove_OutputsDeleted(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf`)

	// Create a fake deployed output
	outputPath := filepath.Join(home, "generated", "config", "fzf.conf")
	require.NoError(t, os.MkdirAll(filepath.Dir(outputPath), 0755))
	require.NoError(t, os.WriteFile(outputPath, []byte("fzf config"), 0644))

	st := state.New()
	st.Presets["tools/fzf"] = state.PresetState{
		Status:  "installed",
		Outputs: []string{outputPath},
	}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	opts := RemoveOptions{
		PresetIDs: []string{"tools/fzf"},
		Yes:       true,
		NoSave:    true,
	}

	err := RunRemove(opts)
	require.NoError(t, err)

	// Output file should be gone
	assert.NoFileExists(t, outputPath)
}

func TestRunRemove_PartialPreset(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/broken", `name: broken`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	st := state.New()
	st.Presets["tools/broken"] = state.PresetState{Status: "partial"}
	require.NoError(t, st.Write(filepath.Join(home, "local", "gop1x.state.yaml")))

	opts := RemoveOptions{
		PresetIDs: []string{"tools/broken"},
		Yes:       true,
		NoSave:    true,
	}

	err := RunRemove(opts)
	require.NoError(t, err)

	st2, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	_, exists := st2.Presets["tools/broken"]
	assert.False(t, exists)
}
