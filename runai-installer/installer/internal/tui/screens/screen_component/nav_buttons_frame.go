package screencomponent

import (
	tea "github.com/charmbracelet/bubbletea"
)

type NavButtonType int

const (
	NavButtonTypeNext NavButtonType = iota
	NavButtonTypeBack
	NavButtonTypeExit
)

type NavButton struct {
	Name  string
	Type  NavButtonType
	Event func() tea.Cmd
}

var NextButton = NavButton{
	Name: "Next",
	Type: NavButtonTypeNext,
	Event: func() tea.Cmd {
		return func() tea.Msg {
			return NextScreenMsg{}
		}
	},
}

var BackButton = NavButton{
	Name: "Back",
	Type: NavButtonTypeBack,
	Event: func() tea.Cmd {
		return func() tea.Msg {
			return BackNavigationMsg{}
		}
	},
}

var ExitButton = NavButton{
	Type: NavButtonTypeExit,
	Name: "Exit",
	Event: func() tea.Cmd {
		return tea.Quit
	},
}

type BackNavigationMsg struct{}
type NextScreenMsg struct{}

func GetNavButtons(isFirstScreen bool, isLastScreen bool) []NavButton {
	navButtons := []NavButton{}
	if isFirstScreen {
		navButtons = []NavButton{NextButton, ExitButton}
	} else if isLastScreen {
		navButtons = []NavButton{BackButton, ExitButton}
	} else {
		navButtons = []NavButton{BackButton, NextButton, ExitButton}
	}
	return navButtons
}

func (n *NavButtons) GetUpdateOrder() int {
	return 1
}
