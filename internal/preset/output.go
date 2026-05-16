package preset

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ohp1x/gop1x/internal/config"
)

type OutputProcessor struct {
	Home              string
	UserHome          string
	XDGConfig         string
	GeneratedDir      string
	LocalGeneratedDir string
}

func NewOutputProcessor() *OutputProcessor {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, _ := os.UserHomeDir()
		xdg = filepath.Join(home, ".config")
	}
	userHome, _ := os.UserHomeDir()

	return &OutputProcessor{
		Home:              config.Home(),
		UserHome:          userHome,
		XDGConfig:         xdg,
		GeneratedDir:      config.GeneratedDir(),
		LocalGeneratedDir: config.LocalGeneratedDir(),
	}
}

func InferMode(o Output) OutputMode {
	if o.Mode != "" {
		return o.Mode
	}
	if strings.HasPrefix(o.Target, "shell/") {
		return OutputConcat
	}
	if strings.HasSuffix(o.Source, ".tmpl") {
		return OutputTemplate
	}
	return OutputCopy
}

func (p *OutputProcessor) ResolveTarget(o Output) (string, error) {
	target := o.Target
	genDir := p.GeneratedDir
	if o.Local {
		genDir = p.LocalGeneratedDir
	}

	switch {
	case strings.HasPrefix(target, "shell/"):
		return filepath.Join(genDir, target), nil
	case strings.HasPrefix(target, "config/"):
		return filepath.Join(genDir, target), nil
	case strings.HasPrefix(target, "home/"):
		rel := strings.TrimPrefix(target, "home/")
		return filepath.Join(p.UserHome, rel), nil
	case strings.HasPrefix(target, "xdg/"):
		rel := strings.TrimPrefix(target, "xdg/")
		return filepath.Join(p.XDGConfig, rel), nil
	default:
		return filepath.Join(genDir, target), nil
	}
}

func (p *OutputProcessor) Deploy(o Output, presetDir string, tplData *TemplateData) (string, error) {
	mode := InferMode(o)

	if mode == OutputConcat {
		return "", nil
	}

	targetPath, err := p.ResolveTarget(o)
	if err != nil {
		return "", err
	}

	switch mode {
	case OutputMkdir:
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return "", fmt.Errorf("mkdir %s: %w", targetPath, err)
		}
		return targetPath, nil

	case OutputCopy:
		srcPath := filepath.Join(presetDir, o.Source)
		if err := copyFile(srcPath, targetPath); err != nil {
			return "", fmt.Errorf("copy %s → %s: %w", srcPath, targetPath, err)
		}
		return targetPath, nil

	case OutputTemplate:
		srcPath := filepath.Join(presetDir, o.Source)
		if err := renderTemplate(srcPath, targetPath, tplData); err != nil {
			return "", fmt.Errorf("template %s → %s: %w", srcPath, targetPath, err)
		}
		return targetPath, nil

	case OutputSymlink:
		srcPath := filepath.Join(presetDir, o.Source)
		if err := createSymlink(srcPath, targetPath); err != nil {
			return "", fmt.Errorf("symlink %s → %s: %w", srcPath, targetPath, err)
		}
		return targetPath, nil
	}

	return "", fmt.Errorf("unknown output mode: %s", mode)
}

func (p *OutputProcessor) DeployAll(outputs []Output, presetDir string, tplData *TemplateData) ([]string, error) {
	var deployed []string
	for _, o := range outputs {
		path, err := p.Deploy(o, presetDir, tplData)
		if err != nil {
			return deployed, err
		}
		if path != "" {
			deployed = append(deployed, path)
		}
	}
	return deployed, nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func renderTemplate(src, dst string, data *TemplateData) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	tmpl, err := template.ParseFiles(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, data)
}

func createSymlink(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	os.Remove(dst)
	return os.Symlink(src, dst)
}
