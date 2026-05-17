package pkg

import (
	"os/exec"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Dnf struct {
	r *iexec.Runner
}

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
	res := d.r.RunElevated("dnf install", "dnf", args...)
	return res.Err
}

func (d *Dnf) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"remove", "-y"}, pkgs...)
	res := d.r.RunElevated("dnf remove", "dnf", args...)
	return res.Err
}

func (d *Dnf) IsInstalled(pkg string) bool {
	res := d.r.Run("rpm", "-q", pkg)
	return res.Err == nil
}
