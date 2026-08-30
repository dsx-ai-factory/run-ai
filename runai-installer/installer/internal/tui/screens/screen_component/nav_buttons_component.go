package screencomponent

import (
	"strings"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	styles "github.com/run-ai/runai/runai-installer/installer/internal/tui/styles"
)

type NavButtons struct {
	viewPort       viewport.Model
	mainScreenData *MultiFrameScreenData
	navButtons     []NavButton
}

func NewNavButtonsFrame(mainScreenData *MultiFrameScreenData) ScreenFrame {
	navButtons := &NavButtons{
		mainScreenData: mainScreenData,
		viewPort:       viewport.New(0, 0),
	}
	navButtons.updateViewPort()
	return navButtons
}

func (n *NavButtons) Event() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return NavButtonsFrameFocusedMsg{}
		}
	}
}

func (n *NavButtons) getFilteredButtons() []NavButton {
	allButtons := GetNavButtons(n.mainScreenData.isFirstScreen(), n.mainScreenData.isLastScreen())
	filtered := []NavButton{}
	for _, button := range allButtons {
		if button.Type == NavButtonTypeNext {
			if !n.mainScreenData.isLastScreen() && !n.mainScreenData.CurrentScreen.CanProceed() {
				continue
			}
		}
		filtered = append(filtered, button)
	}
	return filtered
}

func (n *NavButtons) OnFocus() {
	n.mainScreenData.FocusedScreenFrame = viewscommon.ScreenFrameTypeNavButtons
	n.navButtons = n.getFilteredButtons()

	// Default to first enabled button
	n.mainScreenData.SelectedButton = n.findFirstEnabledButton()

	// Prefer Next button if available
	for i, btn := range n.navButtons {
		if btn.Type == NavButtonTypeNext {
			n.mainScreenData.SelectedButton = i
			break
		}
	}
}

func (n *NavButtons) OnBlur() {
	n.mainScreenData.SelectedButton = -1
}

func (n *NavButtons) isButtonEnabled(buttonIndex int) bool {
	if buttonIndex < 0 || buttonIndex >= len(n.navButtons) {
		return false
	}
	// If the button is present in the list, it is enabled (disabled ones are filtered out)
	return true
}

func (n *NavButtons) findFirstEnabledButton() int {
	for i := range n.navButtons {
		if n.isButtonEnabled(i) {
			return i
		}
	}
	// If no enabled button found, return 0 as fallback
	return 0
}

func (n *NavButtons) SetSize(width int, height int) {
	h := int(0.1 * float64(height))
	n.viewPort.Width, n.viewPort.Height = width, h
}

func (n *NavButtons) Render() string {
	n.updateViewPort()
	return n.viewPort.View()
}

func (n *NavButtons) updateViewPort() {
	var sb strings.Builder
	sb.WriteString("\n")
	previousButtonCount := len(n.navButtons)
	n.navButtons = n.getFilteredButtons()

	// Ensure selected button is valid
	if n.mainScreenData.SelectedButton >= len(n.navButtons) {
		n.mainScreenData.SelectedButton = n.findFirstEnabledButton()
	}

	// If focused on nav buttons and button list changed (Next became available), prefer Next
	if n.mainScreenData.FocusedScreenFrame == viewscommon.ScreenFrameTypeNavButtons && len(n.navButtons) > previousButtonCount {
		for i, btn := range n.navButtons {
			if btn.Type == NavButtonTypeNext {
				n.mainScreenData.SelectedButton = i
				break
			}
		}
	}

	for i, button := range n.navButtons {
		var style lipgloss.Style
		if i == n.mainScreenData.SelectedButton && n.mainScreenData.FocusedScreenFrame == viewscommon.ScreenFrameTypeNavButtons {
			style = styles.AppStyles.ButtonFocused
		} else {
			style = styles.AppStyles.Button
		}
		// Add some padding and make buttons more prominent
		buttonText := " " + button.Name + " "
		sb.WriteString(style.Render(buttonText))
		if i < len(n.navButtons)-1 {
			sb.WriteString("  ") // Add spacing between buttons
		}
	}
	content := sb.String()
	n.viewPort.SetContent(content)
}

func (n *NavButtons) Update(msg tea.Msg) tea.Cmd {
	n.updateViewPort()
	if n.mainScreenData.FocusedScreenFrame != viewscommon.ScreenFrameTypeNavButtons {
		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if n.mainScreenData.SelectedButton >= 0 && n.mainScreenData.SelectedButton < len(n.navButtons) {
				return n.navButtons[n.mainScreenData.SelectedButton].Event()
			}
			return nil
		case tea.KeyRight:
			buttonCount := len(n.navButtons)
			if buttonCount == 0 {
				return nil
			}
			nextButton := (n.mainScreenData.SelectedButton + 1) % buttonCount
			n.mainScreenData.SelectedButton = nextButton
			n.updateViewPort()
		case tea.KeyLeft:
			buttonCount := len(n.navButtons)
			if buttonCount == 0 {
				return nil
			}
			nextButton := buttonCount - (1 - n.mainScreenData.SelectedButton)
			nextButton = nextButton % buttonCount
			n.mainScreenData.SelectedButton = nextButton
			n.updateViewPort()
		case tea.KeyUp:
			if n.mainScreenData.CurrentScreen.GetItemCount() == 0 {
				return nil
			}
			return func() tea.Msg {
				return viewscommon.FocusFrameMsg{
					FrameType: viewscommon.ScreenFrameTypeMain,
					Position:  viewscommon.FocusPositionLast,
				}
			}
		case tea.KeyDown:
			if n.mainScreenData.CurrentScreen.GetItemCount() == 0 {
				return nil
			}
			return func() tea.Msg {
				return viewscommon.FocusFrameMsg{
					FrameType: viewscommon.ScreenFrameTypeMain,
					Position:  viewscommon.FocusPositionFirst,
				}
			}
		}
	}
	return nil
}
