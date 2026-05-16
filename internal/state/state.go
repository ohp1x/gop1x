package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type State struct {
	SchemaVersion int                    `yaml:"schema_version"`
	Presets       map[string]PresetState `yaml:"presets"`
}

type PresetState struct {
	InstalledAt time.Time `yaml:"installed_at"`
	InstalledBy string    `yaml:"installed_by"`
	PresetHash  string    `yaml:"preset_hash"`
	ConfigHash  string    `yaml:"config_hash"`
	Packages    []string  `yaml:"packages"`
	Outputs     []string  `yaml:"outputs"`
	Status      string    `yaml:"status"` // installed, partial, pending_removal
}

func New() *State {
	return &State{
		SchemaVersion: 1,
		Presets:       make(map[string]PresetState),
	}
}

func Read(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	s := New()
	if err := yaml.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	return s, nil
}

func (s *State) Write(path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Atomic write: write to temp file, then rename
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
