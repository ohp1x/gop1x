package preset

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/ui"
)

const (
	DefaultPresetsRepo = "https://github.com/ohp1x/presets.git"
	OfficialPresetsDir = "ohp1x"
)

type PresetsCloner struct {
	LocalPresetsDir string
	PresetsRepo     string
}

func NewPresetsCloner() *PresetsCloner {
	return &PresetsCloner{
		LocalPresetsDir: config.LocalPresetsDir(),
		PresetsRepo:     DefaultPresetsRepo,
	}
}

func (pc *PresetsCloner) EnsureCloned() error {
	officialDir := filepath.Join(pc.LocalPresetsDir, OfficialPresetsDir)

	// Check if already cloned
	if _, err := os.Stat(filepath.Join(officialDir, ".git")); err == nil {
		// Already cloned, check if we need to update
		return pc.updateIfNeeded(officialDir)
	}

	// Clone presets repo
	if err := os.MkdirAll(pc.LocalPresetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create presets dir: %w", err)
	}

	ui.Printf("Cloning presets repository...")
	cmd := exec.Command("git", "clone", pc.PresetsRepo, officialDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone presets: %w", err)
	}

	ui.Ok("Presets repository cloned successfully")
	return nil
}

func (pc *PresetsCloner) updateIfNeeded(officialDir string) error {
	// For now, just verify it exists
	// Future: implement smart update logic (git pull with conflict handling)
	if _, err := os.Stat(officialDir); err != nil {
		return fmt.Errorf("presets directory corrupted: %w", err)
	}
	return nil
}
