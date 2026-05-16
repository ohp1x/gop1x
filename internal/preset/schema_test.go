package preset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected Preset
	}{
		{
			name: "minimal preset",
			yaml: `
name: tools/fzf
description: Fuzzy finder
version: "1.0"
os: [linux, darwin]
packages:
  - fzf
`,
			expected: Preset{
				Name:        "tools/fzf",
				Description: "Fuzzy finder",
				Version:     "1.0",
				OS:          []string{"linux", "darwin"},
				Packages:    []string{"fzf"},
			},
		},
		{
			name: "preset with outputs and hooks",
			yaml: `
name: shell/zsh
description: Zsh configuration
version: "1.0"
depends:
  - core/system
packages:
  - zsh
outputs:
  - target: shell/zshrc.zsh
    source: zshrc.zsh
    priority: 50
hooks:
  post_install: hooks/setup.sh
`,
			expected: Preset{
				Name:        "shell/zsh",
				Description: "Zsh configuration",
				Version:     "1.0",
				Depends:     []string{"core/system"},
				Packages:    []string{"zsh"},
				Outputs: []Output{
					{Target: "shell/zshrc.zsh", Source: "zshrc.zsh", Priority: 50},
				},
				Hooks: Hooks{PostInstall: "hooks/setup.sh"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "preset.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tt.yaml), 0644))

			got, err := Load(path)
			require.NoError(t, err)
			require.Equal(t, tt.expected.Name, got.Name)
			require.Equal(t, tt.expected.Description, got.Description)
			require.Equal(t, tt.expected.Version, got.Version)
			require.Equal(t, tt.expected.Packages, got.Packages)
		})
	}
}
