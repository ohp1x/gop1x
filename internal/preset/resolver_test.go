package preset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockLoader(presets map[string]*Preset) func(string) (*Preset, error) {
	return func(id string) (*Preset, error) {
		p, ok := presets[id]
		if !ok {
			return nil, fmt.Errorf("preset %q not found", id)
		}
		return p, nil
	}
}

func TestResolveOrder_NoDeps(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"tools/fzf": {Name: "fzf"},
	})

	order, err := ResolveOrder("tools/fzf", load)
	require.NoError(t, err)
	assert.Equal(t, []string{"tools/fzf"}, order)
}

func TestResolveOrder_LinearChain(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"tools/fzf":  {Name: "fzf", Depends: []string{"zsh/omz"}},
		"zsh/omz":    {Name: "omz", Depends: []string{"shell/zsh"}},
		"shell/zsh":  {Name: "zsh", Depends: []string{"core/system"}},
		"core/system": {Name: "system"},
	})

	order, err := ResolveOrder("tools/fzf", load)
	require.NoError(t, err)
	assert.Equal(t, []string{"core/system", "shell/zsh", "zsh/omz", "tools/fzf"}, order)
}

func TestResolveOrder_Diamond(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"A": {Name: "A", Depends: []string{"B", "C"}},
		"B": {Name: "B", Depends: []string{"D"}},
		"C": {Name: "C", Depends: []string{"D"}},
		"D": {Name: "D"},
	})

	order, err := ResolveOrder("A", load)
	require.NoError(t, err)
	assert.Equal(t, []string{"D", "B", "C", "A"}, order)
}

func TestResolveOrder_CycleDetection(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"A": {Name: "A", Depends: []string{"B"}},
		"B": {Name: "B", Depends: []string{"A"}},
	})

	_, err := ResolveOrder("A", load)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestResolveOrder_MissingDep(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"A": {Name: "A", Depends: []string{"missing"}},
	})

	_, err := ResolveOrder("A", load)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestResolveOrderMulti_SharedDeps(t *testing.T) {
	load := mockLoader(map[string]*Preset{
		"X": {Name: "X", Depends: []string{"shared"}},
		"Y": {Name: "Y", Depends: []string{"shared"}},
		"shared": {Name: "shared"},
	})

	order, err := ResolveOrderMulti([]string{"X", "Y"}, load)
	require.NoError(t, err)
	assert.Equal(t, []string{"shared", "X", "Y"}, order)
}
