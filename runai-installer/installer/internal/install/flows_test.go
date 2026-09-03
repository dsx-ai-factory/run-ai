/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetFlowsToInstall_Full(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.SelfHosted,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)

	assert.Len(t, flows, 2, "Full mode should return 2 flows")
	assert.Equal(t, "Control Plane", flows[0].GetName())
	assert.Equal(t, "Runai cluster", flows[1].GetName())
}

func TestGetFlowsToInstall_CpOnly(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.CPOnly,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)

	assert.Len(t, flows, 1, "CP-only mode should return 1 flow")
	assert.Equal(t, "Control Plane", flows[0].GetName())
}

func TestGetFlowsToInstall_DefaultConfig(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)

	assert.Len(t, flows, 2, "Default config (CpOnly=false) should return 2 flows")
	assert.Equal(t, "Control Plane", flows[0].GetName())
	assert.Equal(t, "Runai cluster", flows[1].GetName())
}

func TestGetFlowsToInstall_ClusterSteps_PathRouting(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{},
		},
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.SelfHosted,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)
	clusterFlow := flows[1]

	assert.Equal(t, "Runai cluster", clusterFlow.GetName())
	assert.Len(t, clusterFlow.GetSteps(), 5)
}

func TestGetFlowsToInstall_ClusterSteps_HostRouting(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputWorkloadRouting: pkgvalidation.WorkloadRoutingHost,
			},
		},
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.SelfHosted,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)
	clusterFlow := flows[1]

	assert.Equal(t, "Runai cluster", clusterFlow.GetName())
	assert.Len(t, clusterFlow.GetSteps(), 6)
	assert.Equal(t, "Install host-based routing", clusterFlow.GetSteps()[2].GetName())
}

func TestGetFlowsToInstall_ClusterSteps_HostRouting_OpenShift(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		InstallValues: model.InstallValues{
			Inputs: map[configinputs.InstallInputType]string{
				configinputs.InstallInputWorkloadRouting:        "host",
				configinputs.InstallInputKubernetesDistribution: "openshift",
			},
		},
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.SelfHosted,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)
	clusterFlow := flows[1]

	assert.Equal(t, "Runai cluster", clusterFlow.GetName())
	assert.Len(t, clusterFlow.GetSteps(), 5)
	for _, step := range clusterFlow.GetSteps() {
		assert.NotEqual(t, "Install host-based routing", step.GetName())
	}
}

func TestGetFlowsToInstall_Researcher(t *testing.T) {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	runaiData := &model.RunAIInstallData{
		Config: &config.InstallConfig{
			RunaiInstallFormat: config.SHWithResearcherSetup,
		},
	}

	flows := GetFlowsToInstall(mockK8sClient, step.NewHelmInstaller(), nil, runaiData)

	assert.Len(t, flows, 3, "Default config (CpOnly=false) should return 2 flows")
	assert.Equal(t, "Control Plane", flows[0].GetName())
	assert.Equal(t, "Runai cluster", flows[1].GetName())
	assert.Equal(t, "Researcher setup", flows[2].GetName())
}
