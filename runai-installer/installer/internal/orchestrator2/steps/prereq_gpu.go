/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
)

const (
	gpuOperatorRelease   = "gpu-operator"
	gpuOperatorChart     = "gpu-operator"
	gpuOperatorNamespace = "gpu-operator"
	gpuOperatorSentinel  = "gpu-operator"

	ngcHelmRepo = "https://helm.ngc.nvidia.com/nvidia"

	fakeGPURepo  = "oci://ghcr.io/run-ai/fake-gpu-operator/fake-gpu-operator"
	fakeGPULabel = "run.ai/simulated-gpu-node-pool"
	fakeGPUValue = "default"
)

// NewGPUOperatorReal builds the real NVIDIA GPU Operator install step (NGC).
func NewGPUOperatorReal(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     ngcHelmRepo,
		Version:     version,
		ReleaseName: gpuOperatorRelease,
		ChartName:   gpuOperatorChart,
		Namespace:   gpuOperatorNamespace,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-gpu-operator", deploymentSentinel(gpuOperatorSentinel, gpuOperatorNamespace), deps, action)
}

// NewGPUOperatorFake builds the simulated GPU Operator install step: it labels
// worker nodes as simulated GPU nodes, then installs the fake chart. On
// OpenShift it sets environment.openshift=true so the chart binds the privileged
// SCC to the device-plugin/status-exporter ServiceAccounts, whose pods are
// privileged and mount hostPath volumes (rejected by restricted-v2 otherwise).
func NewGPUOperatorFake(version string, customValues map[string]string, distribution string, deps orchestrator2.Deps) orchestrator2.Step {
	label := actions.NodeLabel{Key: fakeGPULabel, Value: fakeGPUValue}
	base := []string{"topology.nodePools.default.gpuCount=8"}
	if distribution == pkgvalidation.KubernetesDistributionOpenShift {
		base = append(base, "environment.openshift=true")
	}
	values := actions.MergeValues(base, customValues)
	install := actions.HelmInstall{
		RepoURL:     fakeGPURepo,
		Version:     version,
		ReleaseName: gpuOperatorRelease,
		ChartName:   gpuOperatorChart,
		Namespace:   gpuOperatorNamespace,
		IsOCI:       true,
		Values:      values,
	}
	return newPrereqStep("install-gpu-operator", deploymentSentinel(gpuOperatorSentinel, gpuOperatorNamespace), deps, label, install)
}
