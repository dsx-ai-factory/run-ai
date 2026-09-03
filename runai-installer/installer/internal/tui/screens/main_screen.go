/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	viewscommon "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
	screencomponent "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/screen_component"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/styles"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

const gap = "\n"

// Model represents the main TUI model
type AppModel struct {
	multiFrameScreenData *screencomponent.MultiFrameScreenData
	screenFrames         []screencomponent.ScreenFrame
	terminalWidth        int
}

// NewModel creates a new TUI model
func NewAppModel(
	runaiData *model.RunAIInstallData,
	multiFrameScreenData *screencomponent.MultiFrameScreenData,
	screenFrames []screencomponent.ScreenFrame,
	k8sClient k8s.K8sClients) *AppModel {
	return &AppModel{
		multiFrameScreenData: multiFrameScreenData,
		screenFrames:         screenFrames,
	}
}

// Init initializes the model
func (m *AppModel) Init() tea.Cmd {
	initCmd := m.multiFrameScreenData.CurrentScreen.Init(m.multiFrameScreenData.RunaiData)

	// After initialization, check if the main frame should be focused
	// If there are no selectable items, focus on nav buttons instead
	if m.multiFrameScreenData.CurrentScreen.GetItemCount() == 0 {
		m.multiFrameScreenData.FocusedScreenFrame = viewscommon.ScreenFrameTypeNavButtons
		m.screenFrames[viewscommon.ScreenFrameTypeNavButtons].OnFocus()
	} else {
		m.screenFrames[viewscommon.ScreenFrameTypeMain].OnFocus()
	}

	// Return batch command: init + request initial window size for proper layout
	return tea.Batch(initCmd, tea.WindowSize())
}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		height := msg.Height - 2*lipgloss.Height(gap)
		for _, component := range m.screenFrames {
			component.SetSize(msg.Width, height)
		}
		// Clear screen on resize to prevent rendering artifacts
		return m, tea.ClearScreen
	case viewscommon.FocusFrameMsg:
		m.screenFrames[m.multiFrameScreenData.FocusedScreenFrame].OnBlur()
		m.multiFrameScreenData.FocusedScreenFrame = msg.FrameType
		m.screenFrames[m.multiFrameScreenData.FocusedScreenFrame].OnFocus()

		if msg.FrameType == viewscommon.ScreenFrameTypeMain {
			switch msg.Position {
			case viewscommon.FocusPositionFirst:
				m.multiFrameScreenData.CurrentScreen.FocusFirst()
			case viewscommon.FocusPositionLast:
				m.multiFrameScreenData.CurrentScreen.FocusLast()
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		default:
			return m, m.updateFrames(msg)
		}
	default:
		return m, m.updateFrames(msg)
	}
}

// View renders the current screen
func (m *AppModel) View() string {
	parts := []string{}
	parts = append(parts, gap)
	title := "NVIDIA Run:ai Installer"
	if m.multiFrameScreenData != nil && m.multiFrameScreenData.K8sClient != nil {
		contextName := m.multiFrameScreenData.K8sClient.GetContextName()
		title = fmt.Sprintf("%s [k8s context: %s]", title, contextName)
	}
	if m.multiFrameScreenData != nil && m.multiFrameScreenData.RunaiData != nil && m.multiFrameScreenData.RunaiData.Config != nil {
		title = fmt.Sprintf("%s \n%+v", title, *m.multiFrameScreenData.RunaiData.Config)
	}
	// Apply width to title so it wraps properly and matches box width
	// Box width = terminal width - 6 (4 padding + 2 border)
	titleStyle := styles.AppStyles.Title
	if m.terminalWidth > 0 {
		titleWidth := m.terminalWidth - 6
		if titleWidth < 10 {
			titleWidth = 10 // Minimum width
		}
		titleStyle = titleStyle.Width(titleWidth)
	}
	parts = append(parts, titleStyle.Render(title))
	for _, frame := range m.screenFrames {
		parts = append(parts, frame.Render())
	}
	parts = append(parts, gap)
	res := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return res
}

func (m *AppModel) updateFrames(msg tea.Msg) tea.Cmd {
	framesCopy := make([]screencomponent.ScreenFrame, len(m.screenFrames))
	copy(framesCopy, m.screenFrames)
	sort.Slice(framesCopy, func(i, j int) bool {
		return framesCopy[i].GetUpdateOrder() < framesCopy[j].GetUpdateOrder()
	})
	for _, frame := range framesCopy {
		cmd := frame.Update(msg)
		if cmd != nil {
			return cmd
		}
	}
	return nil
}
