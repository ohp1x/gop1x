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

func TestCollectRecommends_FiltersInstalled(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
recommends:
  - tools/eza
`)
	writePreset(t, home, "tools/eza", `name: eza
description: modern ls
`)

	locator := preset.NewLocator()
	st := state.New()
	st.Presets["tools/eza"] = state.PresetState{Status: "installed", InstalledBy: "user"}

	entries := collectRecommends([]string{"tools/fzf"}, []string{"tools/fzf"}, locator, st)
	assert.Empty(t, entries)
}

func TestCollectRecommends_FiltersAlreadyInPlan(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
recommends:
  - tools/eza
`)
	writePreset(t, home, "tools/eza", `name: eza
description: modern ls
`)

	locator := preset.NewLocator()
	st := state.New()

	// eza is already in the order (user explicitly added it)
	entries := collectRecommends([]string{"tools/fzf"}, []string{"tools/fzf", "tools/eza"}, locator, st)
	assert.Empty(t, entries)
}

func TestCollectRecommends_CollectsValid(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
recommends:
  - tools/eza
  - tools/bat
`)
	writePreset(t, home, "tools/eza", `name: eza
description: modern ls replacement
`)
	writePreset(t, home, "tools/bat", `name: bat
description: cat with syntax highlighting
`)

	locator := preset.NewLocator()
	st := state.New()

	entries := collectRecommends([]string{"tools/fzf"}, []string{"tools/fzf"}, locator, st)
	assert.Len(t, entries, 2)
	assert.Equal(t, "tools/eza", entries[0].ID)
	assert.Equal(t, "modern ls replacement", entries[0].Description)
	assert.Equal(t, "tools/bat", entries[1].ID)
	assert.Equal(t, "cat with syntax highlighting", entries[1].Description)
}

func TestCollectRecommends_NoTransitive(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
recommends:
  - tools/eza
`)
	writePreset(t, home, "tools/eza", `name: eza
description: modern ls
recommends:
  - tools/bat
`)
	writePreset(t, home, "tools/bat", `name: bat
description: cat with syntax highlighting
`)

	locator := preset.NewLocator()
	st := state.New()

	// Only first-level recommends from rootIDs
	entries := collectRecommends([]string{"tools/fzf"}, []string{"tools/fzf"}, locator, st)
	assert.Len(t, entries, 1)
	assert.Equal(t, "tools/eza", entries[0].ID)
}

func TestRun_YesSkipsRecommends(t *testing.T) {
	home := setupTestEnv(t)

	writePreset(t, home, "tools/fzf", `name: fzf
packages:
  - fzf
recommends:
  - tools/eza
`)
	writePreset(t, home, "tools/eza", `name: eza
description: modern ls
packages:
  - eza
`)
	require.NoError(t, os.WriteFile(filepath.Join(home, "ohp1x.yaml"), []byte("presets: []\n"), 0644))

	pm := &pkg.MockManager{AvailableTrue: true, InstalledMap: map[string]bool{}}

	err := Run(Options{
		PresetIDs:  []string{"tools/fzf"},
		DryRun:     false,
		Yes:        true,
		NoSave:     false,
		PkgManager: pm,
	})
	require.NoError(t, err)

	st, err := state.Read(filepath.Join(home, "local", "gop1x.state.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "installed", st.Presets["tools/fzf"].Status)
	_, hasEza := st.Presets["tools/eza"]
	assert.False(t, hasEza, "recommends should be skipped with --yes")
}
