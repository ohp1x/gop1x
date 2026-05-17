package pkg

import (
	"os"
	"os/exec"
	"runtime"
)

type PackageManager interface {
	Name() string
	IsAvailable() bool
	Install(packages ...string) error
	Uninstall(packages ...string) error
	IsInstalled(pkg string) bool
}

func NewManager() PackageManager {
	switch runtime.GOOS {
	case "darwin":
		return &Brew{}
	case "linux":
		return &Apt{}
	default:
		return nil
	}
}

type Brew struct{}

func (b *Brew) Name() string { return "brew" }

func (b *Brew) IsAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func (b *Brew) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"install"}, pkgs...)
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Brew) IsInstalled(pkg string) bool {
	cmd := exec.Command("brew", "list", pkg)
	return cmd.Run() == nil
}

func (b *Brew) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"uninstall"}, pkgs...)
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Apt struct{}

func (a *Apt) Name() string { return "apt" }

func (a *Apt) IsAvailable() bool {
	_, err := exec.LookPath("apt")
	return err == nil
}

func (a *Apt) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"install", "-y"}, pkgs...)
	cmd := exec.Command("apt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *Apt) IsInstalled(pkg string) bool {
	cmd := exec.Command("dpkg", "-l", pkg)
	return cmd.Run() == nil
}

func (a *Apt) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"remove", "-y"}, pkgs...)
	cmd := exec.Command("apt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
