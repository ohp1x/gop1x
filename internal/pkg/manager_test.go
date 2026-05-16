package pkg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockManager_Install(t *testing.T) {
	m := &MockManager{AvailableTrue: true}

	require.True(t, m.IsAvailable())
	require.Equal(t, "mock", m.Name())

	err := m.Install("fzf", "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"fzf", "ripgrep"}}, m.InstallCalls)
}

func TestMockManager_InstallError(t *testing.T) {
	m := &MockManager{
		AvailableTrue: true,
		InstallErr:    errors.New("install failed"),
	}

	err := m.Install("broken-pkg")
	require.Error(t, err)
	assert.Equal(t, "install failed", err.Error())
}

func TestMockManager_IsInstalled(t *testing.T) {
	m := &MockManager{
		InstalledMap: map[string]bool{"fzf": true},
	}

	assert.True(t, m.IsInstalled("fzf"))
	assert.False(t, m.IsInstalled("ripgrep"))
}
