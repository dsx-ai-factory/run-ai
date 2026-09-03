/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	configinput "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/steps"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type gpuOperator struct {
	*PreReqImpl
}

func NewGpuOperator(k8sClient k8s.K8sClients) *gpuOperator {
	return &gpuOperator{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierGpuOperator, "v25.3.0", true, true, k8sClient),
	}
}

func (g *gpuOperator) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (g *gpuOperator) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (g *gpuOperator) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewGpuOperatorInstallStep(k8sClient, g.GetVersion()),
	}
}

func (g *gpuOperator) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewGpuOperatorInstallStep(g.k8sClient, g.GetVersion()).GetStatusSync(runaiData)
}
