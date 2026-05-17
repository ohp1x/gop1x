package pkg

import (
	"os/exec"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Apt struct {
	r *iexec.Runner
}

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
	res := a.r.RunElevated("apt install", "apt", args...)
	return res.Err
}

func (a *Apt) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"remove", "-y"}, pkgs...)
	res := a.r.RunElevated("apt remove", "apt", args...)
	return res.Err
}

func (a *Apt) IsInstalled(pkg string) bool {
	res := a.r.Run("dpkg", "-l", pkg)
	return res.Err == nil
}
