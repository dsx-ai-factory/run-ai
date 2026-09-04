/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq"
	prereqsteps "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/components"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	viewscommon "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type PrerequisiteInstallScreen struct {
	*viewscommon.InstallationScreen
	runaiData *model.RunAIInstallData
	prereqMap map[string]prereqsteps.Prerequisite
	k8sClient k8s.K8sClients
}

func NewPrerequisiteInstallScreen(k8sClient k8s.K8sClients, prereqMap map[string]prereqsteps.Prerequisite, nextScreen viewscommon.Screen) *PrerequisiteInstallScreen {
	return &PrerequisiteInstallScreen{
		InstallationScreen: viewscommon.NewInstallationScreen("Prerequisite Installation", nextScreen),
		prereqMap:          prereqMap,
		k8sClient:          k8sClient,
	}
}

func (p *PrerequisiteInstallScreen) Init(runaiData *model.RunAIInstallData) tea.Cmd {
	p.runaiData = runaiData
	runaiClientProvider := step.NewRunAIAPIClientProvider(apiclient.NewObjectGetter(p.k8sClient.GetClient()), runaiData.Config.CpVersion)
	flows := prereq.GetFlowsToInstall(p.k8sClient, step.NewHelmInstaller(), runaiClientProvider, runaiData, p.prereqMap)
	p.InstallationScreen.SetFlows(flows)
	return p.InstallationScreen.Init(runaiData)
}
