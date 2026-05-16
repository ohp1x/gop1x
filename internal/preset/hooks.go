package preset

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type HookEnv struct {
	OHP1XHome    string
	PresetID     string
	PresetDir    string
	StateDir     string
	GeneratedDir string
	OS           string
	Arch         string
	PkgManager   string
	Config       map[string]string
}

func RunHook(script string, presetDir string, env *HookEnv) error {
	if script == "" {
		return nil
	}

	scriptPath := script
	if !strings.HasPrefix(script, "/") {
		scriptPath = presetDir + "/" + script
	}

	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("hook script not found: %s", scriptPath)
	}

	cmd := exec.Command("/bin/sh", scriptPath)
	cmd.Dir = presetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = BuildHookEnv(env)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %s failed: %w", script, err)
	}
	return nil
}

func BuildHookEnv(env *HookEnv) []string {
	vars := []string{
		"OHP1X_HOME=" + env.OHP1XHome,
		"GOP1X_PRESET_ID=" + env.PresetID,
		"GOP1X_PRESET_DIR=" + env.PresetDir,
		"GOP1X_STATE_DIR=" + env.StateDir,
		"GOP1X_GENERATED=" + env.GeneratedDir,
		"GOP1X_OS=" + env.OS,
		"GOP1X_ARCH=" + env.Arch,
		"GOP1X_PKG_MANAGER=" + env.PkgManager,
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	for k, v := range env.Config {
		vars = append(vars, "PRESET_CONFIG_"+strings.ToUpper(k)+"="+v)
	}

	return vars
}
