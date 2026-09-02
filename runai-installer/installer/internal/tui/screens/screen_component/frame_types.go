/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screencomponent

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	viewscommon "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
)

type ScreenFrame interface {
	Render() string
	Update(msg tea.Msg) tea.Cmd
	Event() func() tea.Cmd
	OnFocus()
	OnBlur()
	SetSize(width int, height int)
	GetUpdateOrder() int
}

type MainFrameFocusedMsg struct{}
type ErrorsFrameFocusedMsg struct{}
type NavButtonsFrameFocusedMsg struct{}

type MultiFrameScreenData struct {
	FocusedScreenFrame   viewscommon.ScreenFrameType
	ErrorsFrameScrollPos int
	MaxVisibleErrors     int
	CurrentScreen        viewscommon.Screen
	ScreenHistory        []viewscommon.Screen
	SelectedButton       int
	RunaiData            *model.RunAIInstallData
	K8sClient            k8s.K8sClients
}

func NewMultiFrameScreenData(runaiData *model.RunAIInstallData, k8sClient k8s.K8sClients, screen viewscommon.Screen) *MultiFrameScreenData {
	return &MultiFrameScreenData{
		FocusedScreenFrame:   viewscommon.ScreenFrameTypeMain,
		RunaiData:            runaiData,
		K8sClient:            k8sClient,
		ErrorsFrameScrollPos: 0,
		MaxVisibleErrors:     10, // Show up to 10 errors at a time
		CurrentScreen:        screen,
		ScreenHistory:        make([]viewscommon.Screen, 0),
		SelectedButton:       -1,
	}
}

func (m *MultiFrameScreenData) isFirstScreen() bool {
	return len(m.ScreenHistory) == 0
}

func (m *MultiFrameScreenData) isLastScreen() bool {
	return m.CurrentScreen.NextScreen() == nil
}
