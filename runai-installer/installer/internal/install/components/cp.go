/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/install/steps"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type controlplane struct{}

func NewControlPlane() step.InstallationComponent {
	return &controlplane{}
}

func (r *controlplane) GetName() string {
	return "Control Plane"
}

func (r *controlplane) ComponentType() step.ComponentType {
	return step.ComponentTypeControlPlane
}

func (r *controlplane) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	inputs := []configinputs.InstallInputType{
		configinputs.InstallInputFQDN,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
	}
	if !runaiData.Config.DevMode && !runaiData.Config.IsAirgap() {
		inputs = append(inputs, configinputs.InstallInputRegistryPullSecret)
	}
	return inputs
}

func (r *controlplane) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	inputs := []configinputs.InstallInputType{
		configinputs.InstallInputCAChainPath,
	}
	if runaiData.Config.DevMode {
		inputs = append(inputs, configinputs.InstallInputRegistryPullSecret)
	}
	return inputs
}

func (r *controlplane) GetInstallSteps(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}

	createNameSpaceStep := step.NewCreateNamespace(step.CPNamespace, k8sClient)
	installSteps = append(installSteps, createNameSpaceStep)

	if !runaiData.Config.IsAirgap() {
		installSteps = append(installSteps, steps.NewInstallRepoAccessSecret(k8sClient))
	}

	installTlsSecretStep := step.NewInstallTlsSecret(
		steps.RunaiBackendTlsSecretName,
		step.CPNamespace,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
		configinputs.InstallInputFQDN,
		k8sClient,
	)
	installSteps = append(installSteps, installTlsSecretStep)

	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		installSteps = append(installSteps, steps.NewInstallCAChainSecret(steps.RunaiBackendCAChainSecret, step.CPNamespace, step.CPReleaseName, nil, k8sClient))
	}

	if runaiData.Config.IsAirgap() {
		installSteps = append(installSteps, steps.NewInstallAirgapCPChart(helmInstaller, runaiClientProvider))
	} else {
		installSteps = append(installSteps, steps.NewInstallCPChart(helmInstaller, runaiClientProvider))
	}

	return installSteps
}
