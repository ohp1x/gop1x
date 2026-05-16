package pkg

import "runtime"

type PackageManager interface {
	Name() string
	IsAvailable() bool
	Install(packages ...string) error
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

func (b *Brew) Name() string              { return "brew" }
func (b *Brew) IsAvailable() bool          { return false } // TODO: check PATH
func (b *Brew) Install(pkgs ...string) error { return nil } // TODO: implement
func (b *Brew) IsInstalled(pkg string) bool  { return false } // TODO: implement

type Apt struct{}

func (a *Apt) Name() string              { return "apt" }
func (a *Apt) IsAvailable() bool          { return false } // TODO: check PATH
func (a *Apt) Install(pkgs ...string) error { return nil } // TODO: implement
func (a *Apt) IsInstalled(pkg string) bool  { return false } // TODO: implement
