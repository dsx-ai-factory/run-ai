package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all the styling for the TUI
type styles struct {
	Title            lipgloss.Style
	Subtitle         lipgloss.Style
	Normal           lipgloss.Style
	Selected         lipgloss.Style
	Error            lipgloss.Style
	Success          lipgloss.Style
	Warning          lipgloss.Style
	Button           lipgloss.Style
	ButtonFocused    lipgloss.Style
	ButtonDisabled   lipgloss.Style
	Input            lipgloss.Style
	InputFocused     lipgloss.Style
	Checkbox         lipgloss.Style
	CheckboxSelected lipgloss.Style
	Disabled         lipgloss.Style
	MessageBox       lipgloss.Style
	Box              lipgloss.Style
	Placeholder      lipgloss.Style
	Cursor           lipgloss.Style
	// Added colored styles
	ErrorRed     lipgloss.Style
	SuccessGreen lipgloss.Style
}

// Define adaptive palette
var (
	// Grayscale palette
	primary = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#888888"} // mid-gray (darker on light, lighter on dark)

	// Text and backgrounds
	text    = lipgloss.AdaptiveColor{Light: "#111111", Dark: "#F2F2F2"} // near-black / near-white
	subtext = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#BDBDBD"} // medium gray

	// Status colors (grayscale variants)
	errorC  = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"} // dark/light gray
	success = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}
	warning = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#BBBBBB"}

	// Colored status
	red   = lipgloss.Color("#E74C3C")
	green = lipgloss.Color("#2ECC71")

	// Muted neutrals for disabled / placeholder states
	disabledF   = lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#7A7A7A"}
	placeholder = lipgloss.AdaptiveColor{Light: "#8A8A8A", Dark: "#6A6A6A"}

	// Cursor colors
	cursorFG = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}
	cursorBG = lipgloss.AdaptiveColor{Light: "#707070", Dark: "#AFAFAF"}

	// Selected/focused text - high contrast on both backgrounds
	selectedFG = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"} // white on light bg primary, black on dark bg primary
	selectedBG = lipgloss.AdaptiveColor{Light: "#444444", Dark: "#CCCCCC"} // dark gray on light, light gray on dark

	// Button colors - adaptive for both light and dark terminals
	buttonFG         = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#EEEEEE"}
	buttonFocusedFG  = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"}
	buttonFocusedBG  = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}
	buttonDisabledFG = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}
	buttonDisabledBG = lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#3A3A3A"}
)

var AppStyles = &styles{
	Box: lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(false).
		Padding(0, 1),

	MessageBox: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		Padding(0, 2),
	// Width is set dynamically based on terminal width

	Title: lipgloss.NewStyle().
		Foreground(selectedFG).
		Background(selectedBG).
		Padding(0, 1).
		Bold(true),

	Subtitle: lipgloss.NewStyle().
		Foreground(subtext).
		Bold(true),

	Normal: lipgloss.NewStyle().
		Foreground(text),

	Selected: lipgloss.NewStyle().
		Foreground(selectedFG).
		Background(selectedBG).
		Bold(true),

	Error: lipgloss.NewStyle().
		Foreground(errorC),

	Success: lipgloss.NewStyle().
		Foreground(success),

	ErrorRed: lipgloss.NewStyle().
		Foreground(red),

	SuccessGreen: lipgloss.NewStyle().
		Foreground(green),

	Warning: lipgloss.NewStyle().
		Foreground(warning),

	Button: lipgloss.NewStyle().
		Foreground(buttonFG).
		Padding(0, 3).
		Margin(0, 1).
		Bold(true),

	ButtonFocused: lipgloss.NewStyle().
		Foreground(buttonFocusedFG).
		Background(buttonFocusedBG).
		Padding(0, 3).
		Margin(0, 1).
		Bold(true),

	ButtonDisabled: lipgloss.NewStyle().
		Foreground(buttonDisabledFG).
		Background(buttonDisabledBG).
		Padding(0, 3).
		Margin(0, 1),

	Input: lipgloss.NewStyle().
		Foreground(text),

	InputFocused: lipgloss.NewStyle().
		Foreground(selectedFG).
		Background(selectedBG).
		Bold(true),

	Checkbox: lipgloss.NewStyle().
		Foreground(subtext),

	CheckboxSelected: lipgloss.NewStyle().
		Foreground(primary),

	Disabled: lipgloss.NewStyle().
		Foreground(disabledF),

	Placeholder: lipgloss.NewStyle().
		Foreground(placeholder),

	Cursor: lipgloss.NewStyle().
		Foreground(cursorFG).
		Background(cursorBG).
		Underline(true).
		Padding(0, 0),
}
