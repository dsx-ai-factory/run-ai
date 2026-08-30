package ui

import "github.com/charmbracelet/lipgloss"

const (
	CtrlPlusC   = "ctrl+c"
	CtrlPlusR   = "ctrl+r"
	KeyShiftTab = "shift+tab"
	KeyTab      = "tab"
	KeyEsc      = "esc"
	KeyEnter    = "enter"

	KeyUp   = "up"
	KeyDown = "down"

	KeyQ = "q"
	KeyJ = "j"
	KeyK = "k"
)

const (
	PressQToQuit = "\n(press q to quit)\n"

	SelectClusterTableMsg = "\nSelect a %s:\n"

	ErrorNoChoiceWasMade = "no choice was made"
)

// Button types
const (
	ButtonSelectedRound    = "(•) "
	ButtonNotSelectedRound = "( ) "

	ButtonSelectedSquare    = "[X] "
	ButtonNotSelectedSquare = "[ ] "
)

// Colors
const (
	PinkStr  = "205"
	GreenStr = "#00FF00"
	DarkGray = "#3C3C3C"
	Magenta  = "5"
)

var (
	ForegroundPinkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(PinkStr))
	ForegroundGreenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(GreenStr))
	ForegroundMagentaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(Magenta))
)
