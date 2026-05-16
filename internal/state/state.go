package state

import "time"

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
	Status      string    `yaml:"status"`
}

func New() *State {
	return &State{
		SchemaVersion: 1,
		Presets:       make(map[string]PresetState),
	}
}

func Read(path string) (*State, error) {
	// TODO: implement
	return New(), nil
}

func (s *State) Write(path string) error {
	// TODO: implement (atomic write: tmp + rename)
	return nil
}
