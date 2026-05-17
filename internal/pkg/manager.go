package pkg

import (
	"github.com/ohp1x/gop1x/internal/exec"
	"github.com/ohp1x/gop1x/internal/platform"
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
	os := platform.DetectOS()

	switch os {
	case "linux":
		return []PackageManager{
			&Apt{r: r},
			&Dnf{r: r},
			&Pacman{r: r},
			&Apk{r: r},
			&Nix{r: r},
			&Brew{r: r},
		}
	case "darwin":
		return []PackageManager{
			&Brew{r: r},
			&Nix{r: r},
		}
	default:
		return []PackageManager{
			&Brew{r: r},
			&Apt{r: r},
			&Dnf{r: r},
			&Pacman{r: r},
			&Apk{r: r},
			&Nix{r: r},
		}
	}
}
