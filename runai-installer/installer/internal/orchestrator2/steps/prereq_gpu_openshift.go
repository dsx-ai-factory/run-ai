/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	gpuOperatorOCPNamespace = "nvidia-gpu-operator"
	gpuOperatorOCPPackage   = "gpu-operator-certified"
	gpuOperatorOCPSentinel  = "gpu-operator"
	clusterPolicyCRD        = "clusterpolicies.nvidia.com"
	clusterPolicyName       = "gpu-cluster-policy"
	clusterPolicyAPIVersion = "nvidia.com/v1"
	clusterPolicyKind       = "ClusterPolicy"

	clusterMonitoringLabel = "openshift.io/cluster-monitoring"
)

// GPUOpenShiftParams configures the OpenShift GPU operator flow. The OLM
// channels and the ClusterPolicy override are resolved by the planner (from the
// compat matrix and the spec) and passed in, keeping version policy out of this
// assembly code.
type GPUOpenShiftParams struct {
	GPUChannel string
	// GPUVersion is the resolved GPU operator version; it pins the Subscription's
	// startingCSV (gpu-operator-certified.<version>) so the OperatorHub install
	// lands on the same version the Helm path would install.
	GPUVersion string
	NFDChannel string
	// InstallPlanApproval is the OLM upgrade strategy applied to every operator
	// in the flow (NFD + GPU).
	InstallPlanApproval string
	ClusterPolicySpec   map[string]any
	SubscriptionEnv     map[string]string
	GPUCatalog          actions.OLMCatalog
	NFDCatalog          actions.OLMCatalog
}

func NewGPUOperatorOpenShift(p GPUOpenShiftParams, deps orchestrator2.Deps) []orchestrator2.Step {
	out := newNFDSteps(p.NFDChannel, p.InstallPlanApproval, p.NFDCatalog, deps)

	gpuOperator := NewOLMOperator(OLMOperator{
		StepName:            "install-gpu-operator",
		Namespace:           gpuOperatorOCPNamespace,
		NamespaceLabels:     map[string]string{clusterMonitoringLabel: "true"},
		Package:             gpuOperatorOCPPackage,
		Channel:             p.GPUChannel,
		StartingCSV:         startingCSV(gpuOperatorOCPPackage, p.GPUVersion),
		InstallPlanApproval: p.InstallPlanApproval,
		Catalog:             p.GPUCatalog,
		Sentinel:            deploymentSentinel(gpuOperatorOCPSentinel, gpuOperatorOCPNamespace),
		SubscriptionEnv:     p.SubscriptionEnv,
	}, deps)

	clusterPolicy := NewOperatorDefaultCR(OperatorCR{
		StepName:         "create-gpu-cluster-policy",
		CRD:              clusterPolicyCRD,
		CRAPIVersion:     clusterPolicyAPIVersion,
		CRKind:           clusterPolicyKind,
		CRName:           clusterPolicyName,
		CSVNamespace:     gpuOperatorOCPNamespace,
		SubscriptionName: gpuOperatorOCPPackage,
		Override:         p.ClusterPolicySpec,
	}, deps)

	return append(out, gpuOperator, clusterPolicy)
}
