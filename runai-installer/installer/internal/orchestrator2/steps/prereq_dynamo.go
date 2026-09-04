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
	dynamoRelease   = "dynamo-platform"
	dynamoChart     = "dynamo-platform"
	dynamoNamespace = "dynamo-system"
	// dynamoRepo is NGC's classic (ChartMuseum) Helm repository, which serves
	// the chart anonymously. The documented OCI endpoint
	// (oci://helm.ngc.nvidia.com/...) requires NGC credentials and is not used.
	dynamoRepo     = "https://helm.ngc.nvidia.com/nvidia/ai-dynamo"
	dynamoSentinel = "dynamo-platform-dynamo-operator-controller-manager"
)

// NewDynamo builds the NVIDIA Dynamo Platform install step (disaggregated
// inference). The platform chart applies the Dynamo operator's CRDs itself (a
// bundled crds/ dir plus a crd-apply init container), so no separate CRDs chart
// is needed, and it bundles NATS. KAI scheduling and Grove orchestration are
// integrated, not bundled: the Run:ai cluster already provides KAI, and the
// Grove operator is installed by its own feature step (input validation
// enforces that Grove is enabled alongside Dynamo).
func NewDynamo(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	values := actions.MergeValues([]string{
		"global.kai-scheduler.install=false",
		"global.kai-scheduler.enabled=true",
		"global.grove.install=false",
		"global.grove.enabled=true",
	}, customValues)
	action := actions.HelmInstall{
		RepoURL:     dynamoRepo,
		Version:     version,
		ReleaseName: dynamoRelease,
		ChartName:   dynamoChart,
		Namespace:   dynamoNamespace,
		Values:      values,
	}
	return newPrereqStep("install-dynamo", deploymentSentinel(dynamoSentinel, dynamoNamespace), deps, action)
}
