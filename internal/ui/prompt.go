package ui

import (
	"errors"

	"github.com/charmbracelet/huh"
)

var ErrAborted = errors.New("aborted")

func Confirm(prompt string) (bool, error) {
	confirmed := false
	err := huh.NewConfirm().
		Title(prompt).
		Value(&confirmed).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return false, ErrAborted
	}
	if err != nil {
		return false, nil
	}
	return confirmed, nil
}
