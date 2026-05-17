package pkg

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type PackageManager interface {
	Name() string
	IsAvailable() bool
	Install(packages ...string) error
	Uninstall(packages ...string) error
	IsInstalled(pkg string) bool
}

func NewManager() PackageManager {
	candidates := detectCandidates()
	for _, pm := range candidates {
		if pm.IsAvailable() {
			return pm
		}
	}
	return nil
}

func detectCandidates() []PackageManager {
	switch runtime.GOOS {
	case "darwin":
		return []PackageManager{&Brew{}}
	default:
		return []PackageManager{&Apt{}, &Dnf{}, &Pacman{}, &Apk{}, &Nix{}, &Brew{}}
	}
}

// Brew — macOS / Linuxbrew

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

func (b *Brew) IsInstalled(pkg string) bool {
	cmd := exec.Command("brew", "list", pkg)
	return cmd.Run() == nil
}

// Apt — Debian / Ubuntu

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

func (a *Apt) IsInstalled(pkg string) bool {
	cmd := exec.Command("dpkg", "-l", pkg)
	return cmd.Run() == nil
}

// Dnf — Fedora / RHEL

type Dnf struct{}

func (d *Dnf) Name() string { return "dnf" }

func (d *Dnf) IsAvailable() bool {
	_, err := exec.LookPath("dnf")
	return err == nil
}

func (d *Dnf) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"install", "-y"}, pkgs...)
	cmd := exec.Command("dnf", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *Dnf) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"remove", "-y"}, pkgs...)
	cmd := exec.Command("dnf", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *Dnf) IsInstalled(pkg string) bool {
	cmd := exec.Command("rpm", "-q", pkg)
	return cmd.Run() == nil
}

// Pacman — Arch Linux

type Pacman struct{}

func (p *Pacman) Name() string { return "pacman" }

func (p *Pacman) IsAvailable() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (p *Pacman) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-S", "--noconfirm"}, pkgs...)
	cmd := exec.Command("pacman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Pacman) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-Rs", "--noconfirm"}, pkgs...)
	cmd := exec.Command("pacman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Pacman) IsInstalled(pkg string) bool {
	cmd := exec.Command("pacman", "-Qi", pkg)
	return cmd.Run() == nil
}

// Apk — Alpine Linux

type Apk struct{}

func (a *Apk) Name() string { return "apk" }

func (a *Apk) IsAvailable() bool {
	_, err := exec.LookPath("apk")
	return err == nil
}

func (a *Apk) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"add"}, pkgs...)
	cmd := exec.Command("apk", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *Apk) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"del"}, pkgs...)
	cmd := exec.Command("apk", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *Apk) IsInstalled(pkg string) bool {
	cmd := exec.Command("apk", "info", "-e", pkg)
	return cmd.Run() == nil
}

// Nix — NixOS / cross-platform

type Nix struct{}

func (n *Nix) Name() string { return "nix" }

func (n *Nix) IsAvailable() bool {
	_, err := exec.LookPath("nix-env")
	return err == nil
}

func (n *Nix) Install(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	for _, p := range pkgs {
		attr := "nixpkgs." + p
		cmd := exec.Command("nix-env", "-iA", attr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (n *Nix) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-e"}, pkgs...)
	cmd := exec.Command("nix-env", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *Nix) IsInstalled(pkg string) bool {
	out, err := exec.Command("nix-env", "-q", pkg).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), pkg)
}
