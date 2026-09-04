/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type ComponentType int

const (
	ComponentTypeControlPlane ComponentType = iota
	ComponentTypeCluster
	ComponentTypeResearcherSetup
	ComponentTypePrerequisite
	ComponentTypeAirgapPreparation
)

type InstallationComponent interface {
	GetName() string
	ComponentType() ComponentType
	GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType
	GetOptionalInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType
	GetInstallSteps(k8sClient k8s.K8sClients, helmInstaller HelmInstaller, runaiClientProvider RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []InstallStep
}
