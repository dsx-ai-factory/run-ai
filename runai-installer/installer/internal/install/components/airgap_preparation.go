/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/install/steps"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type airgapPreparation struct{}

func NewAirgapPreparation() step.InstallationComponent {
	return &airgapPreparation{}
}

func (a *airgapPreparation) GetName() string {
	return "Airgap Preparation"
}

func (a *airgapPreparation) ComponentType() step.ComponentType {
	return step.ComponentTypeAirgapPreparation
}

func (a *airgapPreparation) GetMandatoryInputs(_ *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputAirgapPackagePath,
		configinputs.InstallInputRegistryURL,
	}
}

func (a *airgapPreparation) GetOptionalInputs(_ *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputRegistryUsername,
		configinputs.InstallInputRegistryPassword,
	}
}

func (a *airgapPreparation) GetInstallSteps(_ k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, _ *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewExtractAirgapPackage(),
		steps.NewPushAirgapImages(),
	}
}
