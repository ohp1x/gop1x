package pkg

import (
	"os/exec"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Apk struct {
	r *iexec.Runner
}

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
	res := a.r.RunElevated("apk add", "apk", args...)
	return res.Err
}

func (a *Apk) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"del"}, pkgs...)
	res := a.r.RunElevated("apk del", "apk", args...)
	return res.Err
}

func (a *Apk) IsInstalled(pkg string) bool {
	res := a.r.Run("apk", "info", "-e", pkg)
	return res.Err == nil
}
