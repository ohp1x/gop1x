package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Compile-time interface compliance checks
var (
	_ PackageManager = (*Brew)(nil)
	_ PackageManager = (*Apt)(nil)
	_ PackageManager = (*Dnf)(nil)
	_ PackageManager = (*Pacman)(nil)
	_ PackageManager = (*Apk)(nil)
	_ PackageManager = (*Nix)(nil)
	_ PackageManager = (*MockManager)(nil)
)

func TestNewManager_ReturnsNonNil(t *testing.T) {
	pm := NewManager()
	assert.NotNil(t, pm)
}

func TestNewManager_DetectsAvailable(t *testing.T) {
	pm := NewManager()
	if pm != nil {
		assert.True(t, pm.IsAvailable())
		assert.NotEmpty(t, pm.Name())
	}
}
