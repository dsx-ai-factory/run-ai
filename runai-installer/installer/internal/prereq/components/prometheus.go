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

type prometheus struct {
	*PreReqImpl
}

func NewPrometheus(k8sClient k8s.K8sClients) *prometheus {
	return &prometheus{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierPrometheus, "76.2.0", true, true, k8sClient),
	}
}

func (p *prometheus) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (p *prometheus) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (p *prometheus) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewPrometheusInstallStep(k8sClient, p.GetVersion()),
	}
}

func (p *prometheus) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewPrometheusInstallStep(p.k8sClient, p.GetVersion()).GetStatusSync(runaiData)
}
