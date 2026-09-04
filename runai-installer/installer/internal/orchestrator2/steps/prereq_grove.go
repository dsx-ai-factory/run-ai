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
	groveRelease   = "grove"
	groveChart     = "grove-charts"
	groveNamespace = "grove-system"
	groveRepo      = "oci://ghcr.io/ai-dynamo/grove/grove-charts"
	groveSentinel  = "grove-operator"
)

// NewGrove builds the Grove operator install step (multi-component AI workload
// orchestration via PodCliqueSets). Grove relies on a PodGang-aware scheduler,
// which the Run:ai cluster already provides through KAI Scheduler.
func NewGrove(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     groveRepo,
		Version:     version,
		ReleaseName: groveRelease,
		ChartName:   groveChart,
		Namespace:   groveNamespace,
		IsOCI:       true,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-grove", deploymentSentinel(groveSentinel, groveNamespace), deps, action)
}
