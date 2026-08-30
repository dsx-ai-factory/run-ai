package common

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ClearScreen returns ANSI escape codes to clear the terminal screen
func ClearScreen() string {
	return "\033[2J\033[H"
}

func HandleListKeys(msg tea.KeyMsg, cursor int, itemCount int) int {
	newCusorLocation := cursor
	switch msg.Type {
	case tea.KeyUp:
		if cursor > 0 {
			newCusorLocation = cursor - 1
		}
	case tea.KeyDown:
		if cursor < itemCount-1 {
			newCusorLocation = cursor + 1
		}
	}
	return newCusorLocation
}

type HandleNavButtonsMsg struct{}
type ScreenSelectionChangeMsg struct{}

func CallNavButtons() tea.Cmd {
	return func() tea.Msg {
		return HandleNavButtonsMsg{}
	}
}

func ScreenSelectionChange() tea.Cmd {
	return func() tea.Msg {
		return ScreenSelectionChangeMsg{}
	}
}
