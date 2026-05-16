package preset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunHook_Success(t *testing.T) {
	presetDir := t.TempDir()
	script := filepath.Join(presetDir, "hook.sh")
	marker := filepath.Join(presetDir, "ran")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0755))

	env := &HookEnv{
		OHP1XHome:    t.TempDir(),
		PresetID:     "test/hook",
		PresetDir:    presetDir,
		StateDir:     t.TempDir(),
		GeneratedDir: t.TempDir(),
		OS:           "linux",
		Arch:         "amd64",
		PkgManager:   "apt",
	}

	err := RunHook("hook.sh", presetDir, env)
	require.NoError(t, err)
	assert.FileExists(t, marker)
}

func TestRunHook_Failure(t *testing.T) {
	presetDir := t.TempDir()
	script := filepath.Join(presetDir, "fail.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nexit 1\n"), 0755))

	env := &HookEnv{
		PresetID:  "test/fail",
		PresetDir: presetDir,
	}

	err := RunHook("fail.sh", presetDir, env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestRunHook_Empty(t *testing.T) {
	err := RunHook("", t.TempDir(), &HookEnv{})
	require.NoError(t, err)
}

func TestRunHook_EnvVars(t *testing.T) {
	presetDir := t.TempDir()
	outFile := filepath.Join(presetDir, "env.txt")
	script := filepath.Join(presetDir, "env.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nenv > "+outFile+"\n"), 0755))

	env := &HookEnv{
		OHP1XHome:    "/home/test/.ohp1x",
		PresetID:     "tools/fzf",
		PresetDir:    presetDir,
		StateDir:     "/home/test/.ohp1x/local",
		GeneratedDir: "/home/test/.ohp1x/generated",
		OS:           "linux",
		Arch:         "amd64",
		PkgManager:   "apt",
		Config:       map[string]string{"theme": "dark"},
	}

	err := RunHook("env.sh", presetDir, env)
	require.NoError(t, err)

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	envStr := string(content)
	assert.Contains(t, envStr, "OHP1X_HOME=/home/test/.ohp1x")
	assert.Contains(t, envStr, "GOP1X_PRESET_ID=tools/fzf")
	assert.Contains(t, envStr, "GOP1X_OS=linux")
	assert.Contains(t, envStr, "PRESET_CONFIG_THEME=dark")
}

func TestBuildHookEnv(t *testing.T) {
	env := &HookEnv{
		OHP1XHome:    "/home/.ohp1x",
		PresetID:     "test/preset",
		PresetDir:    "/presets/test",
		StateDir:     "/state",
		GeneratedDir: "/gen",
		OS:           "darwin",
		Arch:         "arm64",
		PkgManager:   "brew",
		Config:       map[string]string{"editor": "nvim"},
	}

	vars := BuildHookEnv(env)
	assert.Contains(t, vars, "OHP1X_HOME=/home/.ohp1x")
	assert.Contains(t, vars, "GOP1X_PRESET_ID=test/preset")
	assert.Contains(t, vars, "GOP1X_PKG_MANAGER=brew")
	assert.Contains(t, vars, "PRESET_CONFIG_EDITOR=nvim")
}
