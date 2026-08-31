/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package researcher

import (
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/researcher/steps"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type researcherSetup struct{}

func NewResearcherSetup() step.InstallationComponent {
	return &researcherSetup{}
}

func (r *researcherSetup) GetName() string {
	return "Researcher setup"
}

func (r *researcherSetup) ComponentType() step.ComponentType {
	return step.ComponentTypeResearcherSetup
}

func (r *researcherSetup) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputClusterName,
		configinputs.InstallInputResearcherUserEmail,
	}
}

func (r *researcherSetup) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{}
}

func (r *researcherSetup) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}

	userEmail := runaiData.GetInputValue(configinputs.InstallInputResearcherUserEmail)
	if userEmail == "" {
		return installSteps
	}

	k8sAPIClient := k8sClient.GetClient()
	clientsProvider := step.NewRunAIClientsProvider(apiclient.NewObjectGetter(k8sAPIClient), runaiData.Config.CpVersion)

	createUserStep := steps.NewCreateLocalUser(clientsProvider)
	installSteps = append(installSteps, createUserStep)

	getRoleStep := steps.NewGetL1ResearcherRoleID(clientsProvider)
	installSteps = append(installSteps, getRoleStep)

	gpuDiscoverer := steps.NewK8sGPUDiscoverer(k8sAPIClient)
	createProjectStep := steps.NewCreateProject(gpuDiscoverer, clientsProvider, runaiClientProvider)
	installSteps = append(installSteps, createProjectStep)

	createAccessRuleStep := steps.NewCreateAccessRule(createUserStep, getRoleStep, createProjectStep, clientsProvider)
	installSteps = append(installSteps, createAccessRuleStep)

	return installSteps
}
