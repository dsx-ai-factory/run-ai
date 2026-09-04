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

type trainingOperator struct {
	*PreReqImpl
}

func NewTrainingOperator(k8sClient k8s.K8sClients) *trainingOperator {
	return &trainingOperator{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierTrainingOperator, "v1.8.1", false, true, k8sClient),
	}
}

func (t *trainingOperator) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (t *trainingOperator) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (t *trainingOperator) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewTrainingOperatorInstallStep(k8sClient, t.GetVersion()),
	}
}

func (t *trainingOperator) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewTrainingOperatorInstallStep(t.k8sClient, t.GetVersion()).GetStatusSync(runaiData)
}
