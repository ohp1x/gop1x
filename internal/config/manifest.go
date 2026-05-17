package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Shell        string                       `yaml:"shell"`
	Presets      []string                     `yaml:"presets"`
	PresetConfig map[string]map[string]string `yaml:"preset_config"`
}

func ManifestPath() string {
	return filepath.Join(Home(), "ohp1x.yaml")
}

func ReadManifest() (*Manifest, error) {
	data, err := os.ReadFile(ManifestPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{}, nil
		}
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
