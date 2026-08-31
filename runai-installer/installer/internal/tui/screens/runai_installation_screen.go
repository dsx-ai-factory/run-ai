/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/run-ai/runai/runai-installer/installer/internal/install"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	viewscommon "github.com/run-ai/runai/runai-installer/installer/internal/tui/screens/common"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type RunaiInstallationScreen struct {
	*viewscommon.InstallationScreen
	runaiData *model.RunAIInstallData
	k8sClient k8s.K8sClients
}

func NewRunaiInstallationScreen(k8sClient k8s.K8sClients, nextScreen viewscommon.Screen) *RunaiInstallationScreen {
	return &RunaiInstallationScreen{
		InstallationScreen: viewscommon.NewInstallationScreen("Install Run:AI", nextScreen),
		k8sClient:          k8sClient,
	}
}

func (s *RunaiInstallationScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	s.runaiData = runaiData
	runaiClientProvider := step.NewRunAIAPIClientProvider(apiclient.NewObjectGetter(s.k8sClient.GetClient()), runaiData.Config.CpVersion)
	installFlows := install.GetFlowsToInstall(s.k8sClient, step.NewHelmInstaller(), runaiClientProvider, runaiData)
	s.InstallationScreen.SetFlows(installFlows)
	return s.InstallationScreen.Init(runaiData)
}
