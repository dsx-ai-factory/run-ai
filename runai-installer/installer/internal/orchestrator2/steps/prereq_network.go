/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	networkOperatorRelease   = "network-operator"
	networkOperatorChart     = "network-operator"
	networkOperatorNamespace = "nvidia-network-operator"
	networkOperatorSentinel  = "network-operator"
)

// NewNetworkOperator builds the NVIDIA Network Operator install step (NGC).
func NewNetworkOperator(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     ngcHelmRepo,
		Version:     version,
		ReleaseName: networkOperatorRelease,
		ChartName:   networkOperatorChart,
		Namespace:   networkOperatorNamespace,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-network-operator", deploymentSentinel(networkOperatorSentinel, networkOperatorNamespace), deps, action)
}
