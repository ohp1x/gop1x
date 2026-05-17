package ui

import (
	"github.com/charmbracelet/huh"
)

func Confirm(prompt string) bool {
	confirmed := false
	err := huh.NewConfirm().
		Title(prompt).
		Value(&confirmed).
		Run()
	if err != nil {
		return false
	}
	return confirmed
}
