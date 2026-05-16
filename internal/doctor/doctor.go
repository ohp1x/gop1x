package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ohp1x/gop1x/internal/config"
	"github.com/ohp1x/gop1x/internal/pkg"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
)

type Issue struct {
	Level   string // "error", "warning", "info"
	Message string
}

type Report struct {
	Issues []Issue
	OK     bool
}

func Run() *Report {
	report := &Report{
		Issues: []Issue{},
		OK:     true,
	}

	// Check home directory
	checkHomeDir(report)

	// Check state file
	checkStateFile(report)

	// Check package manager
	checkPackageManager(report)

	// Check presets
	checkPresets(report)

	report.OK = !hasErrors(report)
	return report
}

func checkHomeDir(report *Report) {
	homeDir := config.Home()
	if _, err := os.Stat(homeDir); err != nil {
		report.Issues = append(report.Issues, Issue{
			Level:   "error",
			Message: fmt.Sprintf("Home directory not initialized: %s\nRun 'gop1x init' to initialize", homeDir),
		})
		return
	}

	// Check subdirectories
	requiredDirs := []string{
		filepath.Join(homeDir, "presets"),
		filepath.Join(homeDir, "local"),
		filepath.Join(homeDir, "generated"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); err != nil {
			report.Issues = append(report.Issues, Issue{
				Level:   "error",
				Message: fmt.Sprintf("Missing directory: %s", dir),
			})
		}
	}
}

func checkStateFile(report *Report) {
	statePath := config.StatePath()
	s, err := state.Read(statePath)
	if err != nil {
		report.Issues = append(report.Issues, Issue{
			Level:   "error",
			Message: fmt.Sprintf("Failed to read state file: %v", err),
		})
		return
	}

	// Check for partial presets
	for presetID, ps := range s.Presets {
		if ps.Status == "partial" {
			report.Issues = append(report.Issues, Issue{
				Level:   "warning",
				Message: fmt.Sprintf("Preset '%s' has partial status (installation incomplete)", presetID),
			})
		}
	}
}

func checkPackageManager(report *Report) {
	pm := pkg.NewManager()
	if pm == nil {
		report.Issues = append(report.Issues, Issue{
			Level:   "warning",
			Message: fmt.Sprintf("Package manager not available for this OS"),
		})
		return
	}

	if !pm.IsAvailable() {
		report.Issues = append(report.Issues, Issue{
			Level:   "warning",
			Message: fmt.Sprintf("Package manager '%s' not found in PATH", pm.Name()),
		})
	}
}

func checkPresets(report *Report) {
	presetsDir := config.PresetsDir()
	if _, err := os.Stat(presetsDir); err != nil {
		return
	}

	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		report.Issues = append(report.Issues, Issue{
			Level:   "error",
			Message: fmt.Sprintf("Failed to read presets directory: %v", err),
		})
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		presetPath := filepath.Join(presetsDir, entry.Name())
		schemaPath := filepath.Join(presetPath, "preset.yaml")

		if _, err := os.Stat(schemaPath); err != nil {
			report.Issues = append(report.Issues, Issue{
				Level:   "error",
				Message: fmt.Sprintf("Preset '%s' missing preset.yaml", entry.Name()),
			})
			continue
		}

		// Validate preset schema
		_, err := preset.Load(schemaPath)
		if err != nil {
			report.Issues = append(report.Issues, Issue{
				Level:   "error",
				Message: fmt.Sprintf("Preset '%s' invalid: %v", entry.Name(), err),
			})
		}
	}
}

func hasErrors(report *Report) bool {
	for _, issue := range report.Issues {
		if issue.Level == "error" {
			return true
		}
	}
	return false
}
