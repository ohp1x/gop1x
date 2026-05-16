package preset

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ohp1x/gop1x/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectConcatOutputs_SinglePreset(t *testing.T) {
	home := t.TempDir()
	t.Setenv("OHP1X_HOME", home)

	presetDir := filepath.Join(home, "presets", "shell", "zsh")
	require.NoError(t, os.MkdirAll(presetDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(presetDir, "init.zsh"), []byte("# zsh init\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(presetDir, "preset.yaml"), []byte(`name: zsh
outputs:
  - source: init.zsh
    target: shell/zshrc.zsh
    priority: 10
`), 0644))

	st := state.New()
	st.Presets["shell/zsh"] = state.PresetState{Status: "installed"}

	locator := &Locator{UserDir: filepath.Join(home, "presets"), LocalDir: filepath.Join(home, "local", "presets")}
	proc := &OutputProcessor{
		Home:              home,
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	groups, err := CollectConcatOutputs(st, locator, proc)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, filepath.Join(home, "generated", "shell", "zshrc.zsh"), groups[0].Target)
	assert.Len(t, groups[0].Entries, 1)
	assert.Equal(t, "shell/zsh", groups[0].Entries[0].PresetID)
}

func TestCollectConcatOutputs_MultiPresetPriority(t *testing.T) {
	home := t.TempDir()
	t.Setenv("OHP1X_HOME", home)

	for _, tc := range []struct {
		id       string
		content  string
		priority int
	}{
		{"shell/zsh", "# zsh base\n", 10},
		{"tools/fzf", "# fzf bindings\n", 50},
		{"zsh/omz", "# omz init\n", 20},
	} {
		dir := filepath.Join(home, "presets", tc.id)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "init.zsh"), []byte(tc.content), 0644))
		yaml := fmt.Sprintf("name: %s\noutputs:\n  - source: init.zsh\n    target: shell/zshrc.zsh\n    priority: %d\n", tc.id, tc.priority)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "preset.yaml"), []byte(yaml), 0644))
	}

	st := state.New()
	st.Presets["shell/zsh"] = state.PresetState{Status: "installed"}
	st.Presets["tools/fzf"] = state.PresetState{Status: "installed"}
	st.Presets["zsh/omz"] = state.PresetState{Status: "installed"}

	locator := &Locator{UserDir: filepath.Join(home, "presets"), LocalDir: filepath.Join(home, "local", "presets")}
	proc := &OutputProcessor{
		Home:              home,
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	groups, err := CollectConcatOutputs(st, locator, proc)
	require.NoError(t, err)
	require.Len(t, groups, 1)

	entries := groups[0].Entries
	require.Len(t, entries, 3)
	assert.Equal(t, 10, entries[0].Priority)
	assert.Equal(t, 20, entries[1].Priority)
	assert.Equal(t, 50, entries[2].Priority)
}

func TestRegenerateConcat(t *testing.T) {
	home := t.TempDir()

	src1 := filepath.Join(home, "src1.zsh")
	src2 := filepath.Join(home, "src2.zsh")
	require.NoError(t, os.WriteFile(src1, []byte("# first"), 0644))
	require.NoError(t, os.WriteFile(src2, []byte("# second"), 0644))

	target := filepath.Join(home, "generated", "shell", "zshrc.zsh")
	groups := []ConcatGroup{
		{
			Target: target,
			Entries: []ConcatEntry{
				{PresetID: "a", Source: src1, Priority: 10},
				{PresetID: "b", Source: src2, Priority: 20},
			},
		},
	}

	err := RegenerateConcat(groups)
	require.NoError(t, err)

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "# first\n# second\n", string(content))
}
