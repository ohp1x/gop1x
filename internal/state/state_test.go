package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNew(t *testing.T) {
	s := New()
	assert.Equal(t, 1, s.SchemaVersion)
	assert.Empty(t, s.Presets)
}

func TestRead_FileNotFound(t *testing.T) {
	s, err := Read("/tmp/nonexistent-state-file-12345.yaml")
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 1, s.SchemaVersion)
	assert.Empty(t, s.Presets)
}

func TestRead_ValidFile(t *testing.T) {
	tmpfile := t.TempDir() + "/state.yaml"

	// Write a state file
	state := &State{
		SchemaVersion: 1,
		Presets: map[string]PresetState{
			"tools/fzf": {
				InstalledAt: time.Date(2024, 5, 16, 0, 0, 0, 0, time.UTC),
				InstalledBy: "user@example.com",
				PresetHash:  "abc123",
				Status:      "installed",
				Packages:    []string{"fzf"},
			},
		},
	}
	data, err := yaml.Marshal(state)
	require.NoError(t, err)
	err = os.WriteFile(tmpfile, data, 0644)
	require.NoError(t, err)

	// Read it back
	loaded, err := Read(tmpfile)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded.SchemaVersion)
	assert.Len(t, loaded.Presets, 1)
	assert.Contains(t, loaded.Presets, "tools/fzf")
	assert.Equal(t, "installed", loaded.Presets["tools/fzf"].Status)
}

func TestRead_InvalidYAML(t *testing.T) {
	tmpfile := t.TempDir() + "/state.yaml"
	err := os.WriteFile(tmpfile, []byte("invalid: yaml: content:"), 0644)
	require.NoError(t, err)

	_, err = Read(tmpfile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse state file")
}

func TestWrite_CreatesDirectory(t *testing.T) {
	tmpdir := t.TempDir()
	statePath := filepath.Join(tmpdir, "subdir", "state.yaml")

	s := New()
	s.Presets["tools/fzf"] = PresetState{
		Status:   "installed",
		Packages: []string{"fzf"},
	}

	err := s.Write(statePath)
	require.NoError(t, err)

	// Verify file exists
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify no temp file remains
	assert.NoFileExists(t, statePath+".tmp")
}

func TestWrite_AtomicWrite(t *testing.T) {
	tmpdir := t.TempDir()
	statePath := filepath.Join(tmpdir, "state.yaml")

	s := New()
	s.Presets["tools/fzf"] = PresetState{
		Status:   "installed",
		Packages: []string{"fzf"},
	}

	err := s.Write(statePath)
	require.NoError(t, err)

	// Read back
	loaded, err := Read(statePath)
	require.NoError(t, err)
	assert.Len(t, loaded.Presets, 1)
	assert.Equal(t, "installed", loaded.Presets["tools/fzf"].Status)
}

func TestWrite_Overwrites(t *testing.T) {
	tmpdir := t.TempDir()
	statePath := filepath.Join(tmpdir, "state.yaml")

	// First write
	s1 := New()
	s1.Presets["tools/fzf"] = PresetState{Status: "installed"}
	err := s1.Write(statePath)
	require.NoError(t, err)

	// Second write
	s2 := New()
	s2.Presets["tools/ripgrep"] = PresetState{Status: "installed"}
	err = s2.Write(statePath)
	require.NoError(t, err)

	// Verify only second write exists
	loaded, err := Read(statePath)
	require.NoError(t, err)
	assert.Len(t, loaded.Presets, 1)
	assert.Contains(t, loaded.Presets, "tools/ripgrep")
	assert.NotContains(t, loaded.Presets, "tools/fzf")
}

func TestPresetState_Statuses(t *testing.T) {
	tests := []string{"installed", "partial", "pending_removal"}
	for _, status := range tests {
		s := New()
		s.Presets["test/pkg"] = PresetState{Status: status}
		tmpfile := filepath.Join(t.TempDir(), "state.yaml")
		err := s.Write(tmpfile)
		require.NoError(t, err)

		loaded, err := Read(tmpfile)
		require.NoError(t, err)
		assert.Equal(t, status, loaded.Presets["test/pkg"].Status)
	}
}
