package preset

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Preset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Preset
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	return &p, nil
}
