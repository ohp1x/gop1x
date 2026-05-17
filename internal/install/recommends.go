package install

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/ohp1x/gop1x/internal/preset"
	"github.com/ohp1x/gop1x/internal/state"
	"github.com/ohp1x/gop1x/internal/ui"
)

type RecommendEntry struct {
	ID          string
	Description string
}

func collectRecommends(rootIDs []string, order []string, locator *preset.Locator, st *state.State) []RecommendEntry {
	seen := make(map[string]bool)
	for _, id := range order {
		seen[id] = true
	}
	for id := range st.Presets {
		seen[id] = true
	}

	var entries []RecommendEntry
	for _, id := range rootIDs {
		p, err := locator.LoadPreset(id)
		if err != nil {
			continue
		}
		for _, rec := range p.Recommends {
			if seen[rec] {
				continue
			}
			seen[rec] = true
			rp, err := locator.LoadPreset(rec)
			desc := ""
			if err == nil {
				desc = rp.Description
			}
			entries = append(entries, RecommendEntry{ID: rec, Description: desc})
		}
	}
	return entries
}

func promptRecommends(entries []RecommendEntry, yes bool) ([]string, error) {
	if yes || len(entries) == 0 {
		return nil, nil
	}

	if len(entries) == 1 {
		e := entries[0]
		label := e.ID
		if e.Description != "" {
			label = fmt.Sprintf("%s — %s", e.ID, e.Description)
		}
		ui.Printf("\nRecommended (optional):\n  ? %s\n\n", label)
		if ok, err := ui.Confirm("Install recommended?"); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil
		}
		return []string{e.ID}, nil
	}

	ui.Print("\nRecommended (optional):")
	options := make([]huh.Option[string], len(entries))
	for i, e := range entries {
		label := e.ID
		if e.Description != "" {
			label = fmt.Sprintf("%s — %s", e.ID, e.Description)
		}
		options[i] = huh.NewOption(label, e.ID)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select recommended to install").
				Options(options...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ui.ErrAborted
		}
		return nil, nil
	}
	return selected, nil
}
