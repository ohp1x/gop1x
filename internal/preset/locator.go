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
