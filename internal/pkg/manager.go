package pkg

import (
	"github.com/ohp1x/gop1x/internal/exec"
)

type PackageManager interface {
	Name() string
	IsAvailable() bool
	Install(packages ...string) error
	Uninstall(packages ...string) error
	IsInstalled(pkg string) bool
}

func NewManager() PackageManager {
	return NewManagerWithRunner(&exec.Runner{})
}

func NewManagerWithRunner(r *exec.Runner) PackageManager {
	candidates := detectCandidates(r)
	for _, pm := range candidates {
		if pm.IsAvailable() {
			return pm
		}
	}
	return nil
}

func detectCandidates(r *exec.Runner) []PackageManager {
	return []PackageManager{
		&Brew{r: r},
		&Apt{r: r},
		&Dnf{r: r},
		&Pacman{r: r},
		&Apk{r: r},
		&Nix{r: r},
	}
}
