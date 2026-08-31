/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type networkOperator struct {
	*PreReqImpl
}

func NewNetworkOperator(k8sClient k8s.K8sClients) *networkOperator {
	return &networkOperator{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierNetworkOperator, "v25.10.0", true, true, k8sClient),
	}
}

func (n *networkOperator) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (n *networkOperator) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (n *networkOperator) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewNetworkOperatorInstallStep(k8sClient, n.version),
	}
}

func (n *networkOperator) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewNetworkOperatorInstallStep(n.k8sClient, n.GetVersion()).GetStatusSync(runaiData)
}
