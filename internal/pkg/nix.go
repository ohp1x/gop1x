package pkg

import (
	"os/exec"
	"strings"

	iexec "github.com/ohp1x/gop1x/internal/exec"
)

type Nix struct {
	r *iexec.Runner
}

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
		res := n.r.RunWithSpinner("nix install "+p, "nix-env", "-iA", attr)
		if res.Err != nil {
			return res.Err
		}
	}
	return nil
}

func (n *Nix) Uninstall(pkgs ...string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-e"}, pkgs...)
	res := n.r.RunWithSpinner("nix uninstall", "nix-env", args...)
	return res.Err
}

func (n *Nix) IsInstalled(pkg string) bool {
	res := n.r.Run("nix-env", "-q", pkg)
	if res.Err != nil {
		return false
	}
	return strings.Contains(res.Stdout, pkg)
}
