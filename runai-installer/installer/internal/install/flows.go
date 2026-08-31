/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func GetFlowsToInstall(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []*flow.InstallFlow {
	flows := []*flow.InstallFlow{}

	runaiInstallFormat := GetRunaiInstallFormatFromIdentifier(runaiData.Config.RunaiInstallFormat)
	for _, component := range runaiInstallFormat.GetComponents() {
		if runaiData.Config.SkipControlPlane && component.ComponentType() == step.ComponentTypeControlPlane {
			continue
		}
		if runaiData.Config.SkipCluster && component.ComponentType() == step.ComponentTypeCluster {
			continue
		}
		flows = append(flows, flow.NewInstallFlow(component.GetName(), component.GetInstallSteps(k8sClient, helmInstaller, runaiClientProvider, runaiData), runaiData))
	}

	return flows
}
