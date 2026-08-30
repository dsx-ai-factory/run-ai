package screencomponent

import (
	"strings"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	styles "github.com/run-ai/runai/runai-installer/installer/internal/tui/styles"
)

type ErrorsFrame struct {
	viewPort       viewport.Model
	mainScreenData *MultiFrameScreenData
	maxHeight      int
}

func NewErrorsFrame(mainScreenData *MultiFrameScreenData) *ErrorsFrame {
	messages := &ErrorsFrame{
		mainScreenData: mainScreenData,
		viewPort:       viewport.New(0, 0),
		maxHeight:      20, // Default height to prevent 0 constraints before SetSize
	}
	messages.updateViewPort()
	return messages
}

func (m *ErrorsFrame) OnFocus() {
	m.mainScreenData.FocusedScreenFrame = viewscommon.ScreenFrameTypeErrors
}

func (n *ErrorsFrame) OnBlur() {
}

func (m *ErrorsFrame) Event() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return ErrorsFrameFocusedMsg{}
		}
	}
}

func (m *ErrorsFrame) Render() string {
	m.updateViewPort()
	titleStyle := styles.AppStyles.Subtitle
	if m.mainScreenData.FocusedScreenFrame == viewscommon.ScreenFrameTypeErrors {
		titleStyle = styles.AppStyles.Selected
	}
	header := titleStyle.Render(" Errors ")
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewPort.View(),
	)
	// Apply dynamic width based on viewport width, accounting for padding and borders
	// MessageBox has Padding(0, 2) = 4 horizontal padding + 2 border chars = 6 total
	boxWidth := m.viewPort.Width - 6
	if boxWidth < 1 {
		boxWidth = 1 // Minimum width
	}
	return styles.AppStyles.MessageBox.Width(boxWidth).Render(content)
}

func (m *ErrorsFrame) updateViewPort() {
	// Calculate available width for content (viewport width - padding and borders)
	contentWidth := m.viewPort.Width - 6 // 4 horizontal padding + 2 border chars
	if contentWidth < 10 {
		contentWidth = 10 // Minimum width
	}

	// Render messages area (always visible)
	var msgSb strings.Builder
	messages := m.mainScreenData.CurrentScreen.GetMessages()
	if len(messages) > 0 {
		// Wrap each message to fit within the content width
		wrappedMessages := make([]string, 0, len(messages))
		wrapStyle := lipgloss.NewStyle().Width(contentWidth)
		for _, msg := range messages {
			wrappedMessages = append(wrappedMessages, wrapStyle.Render(msg))
		}
		msgSb.WriteString(styles.AppStyles.Normal.Render(strings.Join(wrappedMessages, "\n")))
	} else {
		noMsgText := "No errors."
		if m.mainScreenData.FocusedScreenFrame == viewscommon.ScreenFrameTypeErrors {
			noMsgText += " (Tab to navigate)"
		}
		msgSb.WriteString(styles.AppStyles.Normal.Render(noMsgText))
		// Note: We don't add extra newline here to keep height minimal
	}
	content := msgSb.String()
	m.viewPort.SetContent(content)

	// Calculate required height based on content
	h := lipgloss.Height(content)

	// Constrain height to reasonable portion of screen (e.g. 50%)
	// But ensure at least 1 line is visible
	limit := int(float64(m.maxHeight) * 0.5)
	if limit < 1 {
		limit = 1
	}

	if h > limit {
		h = limit
	}
	if h < 1 {
		h = 1
	}

	m.viewPort.Height = h
}

func (i *ErrorsFrame) SetSize(width int, height int) {
	i.maxHeight = height
	i.viewPort.Width = width
	// Height is determined dynamically in updateViewPort
}

func (i *ErrorsFrame) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	i.updateViewPort()
	if i.mainScreenData.FocusedScreenFrame != viewscommon.ScreenFrameTypeErrors {
		return nil
	}
	i.viewPort, cmd = i.viewPort.Update(msg)
	return cmd
}

func (i *ErrorsFrame) GetUpdateOrder() int {
	return 2
}
