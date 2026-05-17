package pkg

import (
	"os/exec"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Pacman struct {
	r *iexec.Runner
}

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
	res := p.r.RunElevated("pacman install", "pacman", args...)
	return res.Err
}

func (p *Pacman) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-Rs", "--noconfirm"}, pkgs...)
	res := p.r.RunElevated("pacman remove", "pacman", args...)
	return res.Err
}

func (p *Pacman) IsInstalled(pkg string) bool {
	res := p.r.Run("pacman", "-Qi", pkg)
	return res.Err == nil
}
