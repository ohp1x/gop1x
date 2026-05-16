package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHome_UsesEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OHP1X_HOME", dir)

	assert.Equal(t, dir, Home())
}

func TestStatePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OHP1X_HOME", dir)

	expected := filepath.Join(dir, "local", "gop1x.state.yaml")
	assert.Equal(t, expected, StatePath())
}

func TestPresetsDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OHP1X_HOME", dir)

	expected := filepath.Join(dir, "presets")
	assert.Equal(t, expected, PresetsDir())
}
