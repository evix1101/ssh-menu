package ui

import tea "github.com/charmbracelet/bubbletea"

type keyAction int

const (
	keyQuit keyAction = iota
	keySelect
	keyUp
	keyDown
	keyLeft
	keyRight
	keyTab
	keyBackspace
	keyTogglePin
	keyRune
	keyNoop
)

func classifyKey(msg tea.KeyMsg) keyAction {
	switch msg.Type {
	case tea.KeyEscape, tea.KeyCtrlC, tea.KeyCtrlD:
		return keyQuit
	case tea.KeyEnter:
		return keySelect
	case tea.KeyUp:
		return keyUp
	case tea.KeyDown:
		return keyDown
	case tea.KeyLeft:
		return keyLeft
	case tea.KeyRight:
		return keyRight
	case tea.KeyTab:
		return keyTab
	case tea.KeyBackspace:
		return keyBackspace
	case tea.KeyRunes:
		if msg.String() == "p" {
			return keyTogglePin
		}
		return keyRune
	}
	return keyNoop
}
