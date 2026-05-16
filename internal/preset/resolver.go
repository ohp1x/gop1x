package preset

import "fmt"

type color int

const (
	white color = iota
	gray
	black
)

func ResolveOrder(rootID string, load func(id string) (*Preset, error)) ([]string, error) {
	return ResolveOrderMulti([]string{rootID}, load)
}

func ResolveOrderMulti(rootIDs []string, load func(id string) (*Preset, error)) ([]string, error) {
	colors := make(map[string]color)
	var order []string

	var visit func(id string) error
	visit = func(id string) error {
		switch colors[id] {
		case black:
			return nil
		case gray:
			return fmt.Errorf("dependency cycle detected at %q", id)
		}

		colors[id] = gray

		p, err := load(id)
		if err != nil {
			return fmt.Errorf("failed to load preset %q: %w", id, err)
		}

		for _, dep := range p.Depends {
			if err := visit(dep); err != nil {
				return err
			}
		}

		colors[id] = black
		order = append(order, id)
		return nil
	}

	for _, id := range rootIDs {
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	return order, nil
}
