/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/install/steps"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
)

type runaiCluster struct{}

func isOpenShift(runaiData *model.RunAIInstallData) bool {
	return runaiData.GetInputValue(configinputs.InstallInputKubernetesDistribution) == pkgvalidation.KubernetesDistributionOpenShift
}

func needsHostRoutingStep(runaiData *model.RunAIInstallData) bool {
	return runaiData.GetInputValue(configinputs.InstallInputWorkloadRouting) == pkgvalidation.WorkloadRoutingHost && !isOpenShift(runaiData)
}

func NewRunaiCluster() step.InstallationComponent {
	return &runaiCluster{}
}

func (r *runaiCluster) GetName() string {
	return "Runai cluster"
}

func (r *runaiCluster) ComponentType() step.ComponentType {
	return step.ComponentTypeCluster
}

func (r *runaiCluster) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	inputs := []configinputs.InstallInputType{
		configinputs.InstallInputClusterName,
		configinputs.InstallInputFQDN,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
	}
	if !runaiData.Config.DevMode && !runaiData.Config.IsAirgap() {
		inputs = append(inputs, configinputs.InstallInputJfrogCreds)
	}
	if needsHostRoutingStep(runaiData) {
		inputs = append(inputs, configinputs.InstallInputHostRoutingCertPath)
		inputs = append(inputs, configinputs.InstallInputHostRoutingKeyPath)
	}
	return inputs
}

func (r *runaiCluster) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	inputs := []configinputs.InstallInputType{
		configinputs.InstallInputCAChainPath,
		configinputs.InstallInputWorkloadRouting,
	}
	if runaiData.Config.DevMode {
		inputs = append(inputs, configinputs.InstallInputJfrogCreds)
	}
	if !needsHostRoutingStep(runaiData) {
		inputs = append(inputs, configinputs.InstallInputHostRoutingCertPath)
		inputs = append(inputs, configinputs.InstallInputHostRoutingKeyPath)
	}
	return inputs
}

func (r *runaiCluster) GetInstallSteps(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}

	createNameSpaceStep := step.NewCreateNamespace(steps.RunaiClusterNamespace, k8sClient)
	installSteps = append(installSteps, createNameSpaceStep)

	installTlsSecretStep := step.NewInstallTlsSecret(
		steps.RunaiClusterTlsSecretName,
		steps.RunaiClusterNamespace,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
		configinputs.InstallInputFQDN,
		k8sClient,
	)
	installSteps = append(installSteps, installTlsSecretStep)

	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		installSteps = append(installSteps, steps.NewInstallCAChainSecret(steps.RunaiClusterCAChainSecret, steps.RunaiClusterNamespace, step.ClusterReleaseName, steps.GetImmutableRunaiClusterCAChainSecretLablesMap(), k8sClient))
	}

	if needsHostRoutingStep(runaiData) {
		installSteps = append(installSteps, steps.NewInstallHostBasedRouting(k8sClient))
	}

	createClusterStep := steps.NewCreateCluster(runaiClientProvider)
	installSteps = append(installSteps, createClusterStep)

	getClusterInstallationDetailsStep := steps.NewGetClusterInstallationDetails(createClusterStep, runaiClientProvider)
	installSteps = append(installSteps, getClusterInstallationDetailsStep)

	if runaiData.Config.IsAirgap() {
		installSteps = append(installSteps, steps.NewInstallAirgapClusterChart(getClusterInstallationDetailsStep, helmInstaller, runaiClientProvider))
	} else {
		installSteps = append(installSteps, steps.NewInstallClusterChart(getClusterInstallationDetailsStep, helmInstaller, runaiClientProvider))
	}

	return installSteps
}
