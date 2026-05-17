package preset

import (
	"sort"
	"strings"

	"github.com/ohp1x/gop1x/internal/state"
)

type PresetInfo struct {
	Preset    *Preset
	Source    string
	Installed bool
	Status    string
}

func ListPresets(locator *Locator, st *state.State) ([]PresetInfo, error) {
	located, err := locator.ListAll()
	if err != nil {
		return nil, err
	}

	var result []PresetInfo
	for _, lp := range located {
		info := PresetInfo{
			Preset: lp.Preset,
			Source: lp.Source,
		}
		if ps, ok := st.Presets[lp.Preset.Name]; ok {
			info.Installed = ps.Status == "installed"
			info.Status = ps.Status
		}
		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Preset.Name < result[j].Preset.Name
	})

	return result, nil
}

func SearchPresets(locator *Locator, st *state.State, query string) ([]PresetInfo, error) {
	all, err := ListPresets(locator, st)
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)
	var matched []PresetInfo
	for _, info := range all {
		if matchesQuery(info.Preset, q) {
			matched = append(matched, info)
		}
	}
	return matched, nil
}

func matchesQuery(p *Preset, query string) bool {
	if strings.Contains(strings.ToLower(p.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(p.Description), query) {
		return true
	}
	for _, tag := range p.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
