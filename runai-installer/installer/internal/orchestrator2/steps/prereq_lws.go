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
	lwsRelease = "lws"
	lwsChart   = "lws"
	// LWSNamespace is where the upstream LWS Helm chart runs its controller; the
	// Run:ai cluster must be pointed at it.
	LWSNamespace = "lws-system"
	lwsRepo      = "oci://registry.k8s.io/lws/charts/lws"
	lwsSentinel  = "lws-controller-manager"
)

// NewLWS builds the LeaderWorkerSet (distributed inference) install step.
func NewLWS(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     lwsRepo,
		Version:     version,
		ReleaseName: lwsRelease,
		ChartName:   lwsChart,
		Namespace:   LWSNamespace,
		IsOCI:       true,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-lws", deploymentSentinel(lwsSentinel, LWSNamespace), deps, action)
}
