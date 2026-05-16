package preset

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sort"
	"strings"
)

func PresetHash(presetYAMLPath string) (string, error) {
	data, err := os.ReadFile(presetYAMLPath)
	if err != nil {
		return "", fmt.Errorf("hash: read %s: %w", presetYAMLPath, err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:4]), nil
}

func ConfigHash(config map[string]string) (string, error) {
	if len(config) == 0 {
		sum := sha256.Sum256([]byte(""))
		return fmt.Sprintf("%x", sum[:4]), nil
	}

	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(config[k])
		b.WriteByte('\n')
	}

	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", sum[:4]), nil
}

func MergeConfig(presetDefaults map[string]string, overrides map[string]string) map[string]string {
	merged := make(map[string]string, len(presetDefaults)+len(overrides))
	for k, v := range presetDefaults {
		merged[k] = v
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return merged
}
