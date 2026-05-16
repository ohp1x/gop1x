package preset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLocator(t *testing.T) (*Locator, string) {
	home := t.TempDir()
	t.Setenv("OHP1X_HOME", home)

	userDir := filepath.Join(home, "presets")
	localDir := filepath.Join(home, "local", "presets")
	require.NoError(t, os.MkdirAll(userDir, 0755))
	require.NoError(t, os.MkdirAll(localDir, 0755))

	return &Locator{UserDir: userDir, LocalDir: localDir}, home
}

func TestLocator_ResolveUserDir(t *testing.T) {
	loc, _ := setupLocator(t)

	presetDir := filepath.Join(loc.UserDir, "tools", "fzf")
	require.NoError(t, os.MkdirAll(presetDir, 0755))

	path, err := loc.Resolve("tools/fzf")
	require.NoError(t, err)
	assert.Equal(t, presetDir, path)
}

func TestLocator_ResolveLocalDir(t *testing.T) {
	loc, _ := setupLocator(t)

	presetDir := filepath.Join(loc.LocalDir, "tools", "ripgrep")
	require.NoError(t, os.MkdirAll(presetDir, 0755))

	path, err := loc.Resolve("tools/ripgrep")
	require.NoError(t, err)
	assert.Equal(t, presetDir, path)
}

func TestLocator_UserDirPriority(t *testing.T) {
	loc, _ := setupLocator(t)

	userPath := filepath.Join(loc.UserDir, "tools", "fzf")
	localPath := filepath.Join(loc.LocalDir, "tools", "fzf")
	require.NoError(t, os.MkdirAll(userPath, 0755))
	require.NoError(t, os.MkdirAll(localPath, 0755))

	path, err := loc.Resolve("tools/fzf")
	require.NoError(t, err)
	assert.Equal(t, userPath, path)
}

func TestLocator_NotFound(t *testing.T) {
	loc, _ := setupLocator(t)

	_, err := loc.Resolve("nonexistent/preset")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLocator_LoadPreset(t *testing.T) {
	loc, _ := setupLocator(t)

	presetDir := filepath.Join(loc.UserDir, "tools", "fzf")
	require.NoError(t, os.MkdirAll(presetDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(presetDir, "preset.yaml"), []byte("name: fzf\npackages:\n  - fzf\n"), 0644))

	p, err := loc.LoadPreset("tools/fzf")
	require.NoError(t, err)
	assert.Equal(t, "fzf", p.Name)
	assert.Equal(t, []string{"fzf"}, p.Packages)
}
