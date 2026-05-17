package preset

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
)

type Locator struct {
	UserDir  string
	LocalDir string
}

func NewLocator() *Locator {
	return &Locator{
		UserDir:  config.PresetsDir(),
		LocalDir: config.LocalPresetsDir(),
	}
}

func (l *Locator) Resolve(id string) (string, error) {
	userPath := filepath.Join(l.UserDir, id)
	if info, err := os.Stat(userPath); err == nil && info.IsDir() {
		return userPath, nil
	}

	localPath := filepath.Join(l.LocalDir, id)
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		return localPath, nil
	}

	return "", fmt.Errorf("preset %q not found in %s or %s", id, l.UserDir, l.LocalDir)
}

func (l *Locator) LoadPreset(id string) (*Preset, error) {
	dir, err := l.Resolve(id)
	if err != nil {
		return nil, err
	}
	return Load(filepath.Join(dir, "preset.yaml"))
}

type LocatedPreset struct {
	Preset *Preset
	Source string // "user" or "official"
}

func (l *Locator) ListAll() ([]LocatedPreset, error) {
	seen := make(map[string]bool)
	var result []LocatedPreset

	for _, entry := range []struct {
		dir    string
		source string
	}{
		{l.UserDir, "user"},
		{l.LocalDir, "official"},
	} {
		presets, err := walkPresets(entry.dir, entry.source)
		if err != nil {
			continue
		}
		for _, lp := range presets {
			if !seen[lp.Preset.Name] {
				seen[lp.Preset.Name] = true
				result = append(result, lp)
			}
		}
	}

	return result, nil
}

func walkPresets(root, source string) ([]LocatedPreset, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	var result []LocatedPreset

	categories, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, cat := range categories {
		if !cat.IsDir() {
			continue
		}
		catPath := filepath.Join(root, cat.Name())
		names, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}
		for _, name := range names {
			if !name.IsDir() {
				continue
			}
			presetYAML := filepath.Join(catPath, name.Name(), "preset.yaml")
			p, err := Load(presetYAML)
			if err != nil {
				continue
			}
			if p.Name == "" {
				p.Name = cat.Name() + "/" + name.Name()
			}
			result = append(result, LocatedPreset{Preset: p, Source: source})
		}
	}

	return result, nil
}
