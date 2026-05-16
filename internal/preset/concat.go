package preset

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ohp1x/gop1x/internal/state"
)

type ConcatEntry struct {
	PresetID string
	Source   string
	Priority int
	Local    bool
}

type ConcatGroup struct {
	Target  string
	Entries []ConcatEntry
}

func CollectConcatOutputs(st *state.State, locator *Locator, proc *OutputProcessor) ([]ConcatGroup, error) {
	groups := make(map[string]*ConcatGroup)

	for id, ps := range st.Presets {
		if ps.Status != "installed" {
			continue
		}

		dir, err := locator.Resolve(id)
		if err != nil {
			continue
		}

		p, err := Load(filepath.Join(dir, "preset.yaml"))
		if err != nil {
			continue
		}

		for _, o := range p.Outputs {
			mode := InferMode(o)
			if mode != OutputConcat {
				continue
			}

			target, err := proc.ResolveTarget(o)
			if err != nil {
				continue
			}

			entry := ConcatEntry{
				PresetID: id,
				Source:   filepath.Join(dir, o.Source),
				Priority: o.Priority,
				Local:    o.Local,
			}

			if g, ok := groups[target]; ok {
				g.Entries = append(g.Entries, entry)
			} else {
				groups[target] = &ConcatGroup{
					Target:  target,
					Entries: []ConcatEntry{entry},
				}
			}
		}
	}

	result := make([]ConcatGroup, 0, len(groups))
	for _, g := range groups {
		sort.Slice(g.Entries, func(i, j int) bool {
			return g.Entries[i].Priority < g.Entries[j].Priority
		})
		result = append(result, *g)
	}
	return result, nil
}

func RegenerateConcat(groups []ConcatGroup) error {
	for _, g := range groups {
		if err := os.MkdirAll(filepath.Dir(g.Target), 0755); err != nil {
			return fmt.Errorf("concat mkdir %s: %w", g.Target, err)
		}

		out, err := os.Create(g.Target)
		if err != nil {
			return fmt.Errorf("concat create %s: %w", g.Target, err)
		}

		for _, entry := range g.Entries {
			data, err := os.ReadFile(entry.Source)
			if err != nil {
				out.Close()
				return fmt.Errorf("concat read %s: %w", entry.Source, err)
			}
			if _, err := out.Write(data); err != nil {
				out.Close()
				return fmt.Errorf("concat write %s: %w", g.Target, err)
			}
			out.Write([]byte("\n"))
		}
		out.Close()
	}
	return nil
}
