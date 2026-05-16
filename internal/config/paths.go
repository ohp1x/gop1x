package config

import (
	"os"
	"path/filepath"
)

func Home() string {
	if h := os.Getenv("OHP1X_HOME"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ohp1x")
}

func StatePath() string {
	return filepath.Join(Home(), "local", "gop1x.state.yaml")
}

func PresetsDir() string {
	return filepath.Join(Home(), "presets")
}

func LocalPresetsDir() string {
	return filepath.Join(Home(), "local", "presets")
}

func GeneratedDir() string {
	return filepath.Join(Home(), "generated")
}

func LocalGeneratedDir() string {
	return filepath.Join(Home(), "local", "generated")
}
