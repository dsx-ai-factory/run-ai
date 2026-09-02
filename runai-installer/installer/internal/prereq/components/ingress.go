/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinput "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/steps"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type ngnixIngress struct {
	*PreReqImpl
}

func NewNgnixIngress(k8sClient k8s.K8sClients) *ngnixIngress {
	return &ngnixIngress{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierIngress, "4.13.0", true, true, k8sClient),
	}
}

func (n *ngnixIngress) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (n *ngnixIngress) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{
		configinput.InstallInputIngressControllerHTTPSPort,
	}
}

func (n *ngnixIngress) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewNgnixIngressInstallStep(k8sClient, n.version),
	}
}

func (n *ngnixIngress) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewNgnixIngressInstallStep(n.k8sClient, n.GetVersion()).GetStatusSync(runaiData)
}
