package pkg

import (
	"os/exec"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Brew struct {
	r *iexec.Runner
}

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
	res := b.r.RunWithSpinner("brew install", "brew", args...)
	return res.Err
}

func (b *Brew) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"uninstall"}, pkgs...)
	res := b.r.RunWithSpinner("brew uninstall", "brew", args...)
	return res.Err
}

func (b *Brew) IsInstalled(pkg string) bool {
	res := b.r.Run("brew", "list", pkg)
	return res.Err == nil
}
