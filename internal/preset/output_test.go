package preset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInferMode(t *testing.T) {
	tests := []struct {
		name     string
		output   Output
		expected OutputMode
	}{
		{"explicit mode", Output{Mode: OutputSymlink}, OutputSymlink},
		{"shell prefix", Output{Target: "shell/zshrc.zsh"}, OutputConcat},
		{"tmpl suffix", Output{Source: "config.toml.tmpl", Target: "config/app.toml"}, OutputTemplate},
		{"default copy", Output{Source: "gitconfig", Target: "home/.gitconfig"}, OutputCopy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, InferMode(tt.output))
		})
	}
}

func TestResolveTarget(t *testing.T) {
	home := t.TempDir()
	userHome := t.TempDir()
	xdg := t.TempDir()

	proc := &OutputProcessor{
		Home:              home,
		UserHome:          userHome,
		XDGConfig:         xdg,
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	tests := []struct {
		name     string
		output   Output
		expected string
	}{
		{"shell prefix", Output{Target: "shell/zshrc.zsh"}, filepath.Join(home, "generated", "shell", "zshrc.zsh")},
		{"config prefix", Output{Target: "config/starship.toml"}, filepath.Join(home, "generated", "config", "starship.toml")},
		{"home prefix", Output{Target: "home/.gitconfig"}, filepath.Join(userHome, ".gitconfig")},
		{"xdg prefix", Output{Target: "xdg/nvim/init.lua"}, filepath.Join(xdg, "nvim", "init.lua")},
		{"local shell", Output{Target: "shell/local.zsh", Local: true}, filepath.Join(home, "local", "generated", "shell", "local.zsh")},
		{"no prefix", Output{Target: "other/file.txt"}, filepath.Join(home, "generated", "other", "file.txt")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := proc.ResolveTarget(tt.output)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestDeploy_Copy(t *testing.T) {
	home := t.TempDir()
	presetDir := t.TempDir()

	proc := &OutputProcessor{
		Home:              home,
		UserHome:          t.TempDir(),
		XDGConfig:         t.TempDir(),
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	require.NoError(t, os.WriteFile(filepath.Join(presetDir, "gitconfig"), []byte("[user]\nname = test\n"), 0644))

	o := Output{Source: "gitconfig", Target: "config/gitconfig", Mode: OutputCopy}
	path, err := proc.Deploy(o, presetDir, nil)
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "[user]\nname = test\n", string(content))
}

func TestDeploy_Template(t *testing.T) {
	home := t.TempDir()
	presetDir := t.TempDir()

	proc := &OutputProcessor{
		Home:              home,
		UserHome:          t.TempDir(),
		XDGConfig:         t.TempDir(),
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	require.NoError(t, os.WriteFile(filepath.Join(presetDir, "config.tmpl"), []byte("user={{.Config.name}}"), 0644))

	o := Output{Source: "config.tmpl", Target: "config/app.conf", Mode: OutputTemplate}
	tplData := &TemplateData{Config: map[string]string{"name": "alice"}}
	path, err := proc.Deploy(o, presetDir, tplData)
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "user=alice", string(content))
}

func TestDeploy_Symlink(t *testing.T) {
	home := t.TempDir()
	presetDir := t.TempDir()

	proc := &OutputProcessor{
		Home:              home,
		UserHome:          t.TempDir(),
		XDGConfig:         t.TempDir(),
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	srcFile := filepath.Join(presetDir, "vimrc")
	require.NoError(t, os.WriteFile(srcFile, []byte("set nocompatible"), 0644))

	o := Output{Source: "vimrc", Target: "config/vimrc", Mode: OutputSymlink}
	path, err := proc.Deploy(o, presetDir, nil)
	require.NoError(t, err)

	link, err := os.Readlink(path)
	require.NoError(t, err)
	assert.Equal(t, srcFile, link)
}

func TestDeploy_Mkdir(t *testing.T) {
	home := t.TempDir()

	proc := &OutputProcessor{
		Home:              home,
		UserHome:          t.TempDir(),
		XDGConfig:         t.TempDir(),
		GeneratedDir:      filepath.Join(home, "generated"),
		LocalGeneratedDir: filepath.Join(home, "local", "generated"),
	}

	o := Output{Target: "config/nvim", Mode: OutputMkdir}
	path, err := proc.Deploy(o, "", nil)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestDeploy_ConcatSkipped(t *testing.T) {
	proc := &OutputProcessor{
		GeneratedDir:      t.TempDir(),
		LocalGeneratedDir: t.TempDir(),
	}

	o := Output{Target: "shell/zshrc.zsh", Source: "init.zsh"}
	path, err := proc.Deploy(o, t.TempDir(), nil)
	require.NoError(t, err)
	assert.Empty(t, path)
}
