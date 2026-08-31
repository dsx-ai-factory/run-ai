/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/tui/screens"
	screenscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	frames "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/screen_component"
)

func Run(runaiData *model.RunAIInstallData, k8sClients k8s.K8sClients) error {
	model := initModel(runaiData, k8sClients)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("An error occurred: %v", err)
		os.Exit(1)
	}

	return nil
}

func initModel(runaiData *model.RunAIInstallData, k8sClient k8s.K8sClients) tea.Model {
	appScreens := screens.GetScreensWithK8s(k8sClient)
	multiFrameScreenData := frames.NewMultiFrameScreenData(runaiData, k8sClient, appScreens[screenscommon.ScreenPrerequisites])
	screenFrames := make([]frames.ScreenFrame, 3)
	screenFrames[screenscommon.ScreenFrameTypeMain] = frames.NewMainFrame(multiFrameScreenData)
	screenFrames[screenscommon.ScreenFrameTypeErrors] = frames.NewErrorsFrame(multiFrameScreenData)
	screenFrames[screenscommon.ScreenFrameTypeNavButtons] = frames.NewNavButtonsFrame(multiFrameScreenData)
	return screens.NewAppModel(runaiData, multiFrameScreenData, screenFrames, k8sClient)
}
