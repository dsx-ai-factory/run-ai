/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screencomponent

import (
	"strings"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	viewscommon "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
	styles "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/styles"
)

type MainFrame struct {
	mainScreenData    *MultiFrameScreenData
	shouldClearScreen bool
	viewPort          viewport.Model
}

func NewMainFrame(mainScreenData *MultiFrameScreenData) *MainFrame {
	innerView := &MainFrame{
		mainScreenData: mainScreenData,
		viewPort:       viewport.New(0, 0),
	}
	innerView.updateViewPort()
	return innerView

}

func (i *MainFrame) Render() string {
	i.updateViewPort()

	titleStyle := styles.AppStyles.Subtitle
	if i.mainScreenData.FocusedScreenFrame == viewscommon.ScreenFrameTypeMain {
		titleStyle = styles.AppStyles.Selected
	}
	header := titleStyle.Render(i.mainScreenData.CurrentScreen.GetHeader())

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"\n",
		i.viewPort.View(),
	)
	// Apply dynamic width based on viewport width, accounting for padding and borders
	// MessageBox has Padding(0, 2) = 4 horizontal padding + 2 border chars = 6 total
	boxWidth := i.viewPort.Width - 6
	if boxWidth < 1 {
		boxWidth = 1 // Minimum width
	}
	return styles.AppStyles.MessageBox.Width(boxWidth).Render(content)

}

func (i *MainFrame) updateViewPort() {
	var sb strings.Builder

	// Clear screen only when transitioning to a new screen
	if i.shouldClearScreen {
		sb.WriteString(viewscommon.ClearScreen())
		i.shouldClearScreen = false
	}
	i.viewPort.SetContent(i.mainScreenData.CurrentScreen.Render())
}

func (n *MainFrame) OnFocus() {
	n.mainScreenData.FocusedScreenFrame = viewscommon.ScreenFrameTypeMain
	if n.mainScreenData.CurrentScreen != nil {
		n.mainScreenData.CurrentScreen.OnFocus()
	}
}

func (n *MainFrame) OnBlur() {
	if n.mainScreenData.CurrentScreen != nil {
		n.mainScreenData.CurrentScreen.OnBlur()
	}
}

func (i *MainFrame) Event() func() tea.Cmd {
	return func() tea.Cmd {
		return func() tea.Msg {
			return MainFrameFocusedMsg{}
		}
	}
}

func (i *MainFrame) SetSize(width int, height int) {
	h := int(0.3 * float64(height))
	i.viewPort.Width, i.viewPort.Height = width, h
}

func (i *MainFrame) Update(msg tea.Msg) tea.Cmd {
	i.updateViewPort()

	if _, ok := msg.(tea.KeyMsg); ok && i.mainScreenData.FocusedScreenFrame != viewscommon.ScreenFrameTypeMain {
		return nil
	}

	switch msg.(type) {
	case NextScreenMsg:
		i.mainScreenData.CurrentScreen.UpdateSelection(i.mainScreenData.RunaiData)
		nextScreen := i.mainScreenData.CurrentScreen.NextScreen()
		if nextScreen == nil {
			return nil
		}
		i.mainScreenData.ScreenHistory = append(i.mainScreenData.ScreenHistory, i.mainScreenData.CurrentScreen)
		i.mainScreenData.CurrentScreen = nextScreen
		i.shouldClearScreen = true
		initCmd := i.mainScreenData.CurrentScreen.Init(i.mainScreenData.RunaiData)

		targetFrame := viewscommon.ScreenFrameTypeMain
		if i.mainScreenData.CurrentScreen.GetItemCount() == 0 {
			targetFrame = viewscommon.ScreenFrameTypeNavButtons
		}
		focusCmd := func() tea.Msg { return viewscommon.FocusFrameMsg{FrameType: targetFrame} }

		i.updateViewPort()
		return tea.Batch(initCmd, focusCmd)
	case BackNavigationMsg:
		if len(i.mainScreenData.ScreenHistory) > 0 {
			previousScreen := i.mainScreenData.ScreenHistory[len(i.mainScreenData.ScreenHistory)-1]
			i.mainScreenData.ScreenHistory = i.mainScreenData.ScreenHistory[:len(i.mainScreenData.ScreenHistory)-1]
			i.mainScreenData.CurrentScreen = previousScreen

			targetFrame := viewscommon.ScreenFrameTypeMain
			if i.mainScreenData.CurrentScreen.GetItemCount() == 0 {
				targetFrame = viewscommon.ScreenFrameTypeNavButtons
			}
			focusCmd := func() tea.Msg { return viewscommon.FocusFrameMsg{FrameType: targetFrame} }

			return focusCmd
		}
		return nil
	}
	cmd := i.mainScreenData.CurrentScreen.Update(msg)
	if cmd != nil {
		return cmd
	}
	return nil
}

func (i *MainFrame) GetUpdateOrder() int {
	return 0
}
