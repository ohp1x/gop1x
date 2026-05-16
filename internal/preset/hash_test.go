package preset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresetHash_Deterministic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preset.yaml")
	require.NoError(t, os.WriteFile(path, []byte("name: fzf\npackages:\n  - fzf\n"), 0644))

	h1, err := PresetHash(path)
	require.NoError(t, err)
	h2, err := PresetHash(path)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 8)
}

func TestPresetHash_ContentSensitive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preset.yaml")

	require.NoError(t, os.WriteFile(path, []byte("name: fzf\n"), 0644))
	h1, err := PresetHash(path)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(path, []byte("name: ripgrep\n"), 0644))
	h2, err := PresetHash(path)
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2)
}

func TestConfigHash_Deterministic(t *testing.T) {
	cfg := map[string]string{"b": "2", "a": "1", "c": "3"}

	h1, err := ConfigHash(cfg)
	require.NoError(t, err)
	h2, err := ConfigHash(cfg)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 8)
}

func TestConfigHash_OrderIndependent(t *testing.T) {
	cfg1 := map[string]string{"a": "1", "b": "2"}
	cfg2 := map[string]string{"b": "2", "a": "1"}

	h1, err := ConfigHash(cfg1)
	require.NoError(t, err)
	h2, err := ConfigHash(cfg2)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
}

func TestConfigHash_Empty(t *testing.T) {
	h, err := ConfigHash(map[string]string{})
	require.NoError(t, err)
	assert.Len(t, h, 8)
}

func TestMergeConfig(t *testing.T) {
	defaults := map[string]string{"theme": "dark", "editor": "vim"}
	overrides := map[string]string{"editor": "nvim", "shell": "zsh"}

	merged := MergeConfig(defaults, overrides)
	assert.Equal(t, "dark", merged["theme"])
	assert.Equal(t, "nvim", merged["editor"])
	assert.Equal(t, "zsh", merged["shell"])
}
