/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prereq

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func GetFlowsToInstall(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData, preReqAssets map[string]components.Prerequisite) []*flow.InstallFlow {
	flows := []*flow.InstallFlow{}
	configuredPreReq := runaiData.Prerequisites

	for _, preReqIdentifier := range configuredPreReq {
		preReqObj := preReqAssets[preReqIdentifier]
		flows = append(flows, flow.NewInstallFlow(preReqIdentifier, preReqObj.GetInstallSteps(k8sClient, helmInstaller, runaiClientProvider, runaiData), runaiData))
	}
	return flows
}
