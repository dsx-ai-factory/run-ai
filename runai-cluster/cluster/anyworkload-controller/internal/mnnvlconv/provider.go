/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package mnnvlconv provides the Run:AI pod-affinity grouping the dromos mnnvl
// plugin does not own: group GPU pods by run.ai/top-owner-uid, skipped under a
// kai.scheduler/* topology annotation. Internal to anyworkload-controller.
package mnnvlconv

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// kaiTopologyAnnotationKeys must stay byte-identical to kai.scheduler's published
// topology annotation keys.
var kaiTopologyAnnotationKeys = []string{
	"kai.scheduler/topology",
	"kai.scheduler/topology-preferred-placement",
	"kai.scheduler/topology-required-placement",
}

// ProvidePodAffinity groups GPU pods by run.ai/top-owner-uid, or nil when a
// kai.scheduler/* annotation means KAI owns placement. The selector value must
// match what the external-workload-integrator stamps (see topOwnerWorkloadId).
var ProvidePodAffinity mnnvl.PodAffinityProvider = func(
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
) *corev1.PodAffinity {
	if hasAnyKaiTopologyAnnotation(resource) {
		return nil
	}

	return mnnvl.BuildGroupingAffinity(
		commonconstants.TopOwnerUidLabel,
		topOwnerWorkloadId(runner),
		gpuNetworkAccelerationLabelKeys(runner),
	)
}

// topOwnerWorkloadId mirrors what the external-workload-integrator stamps as
// run.ai/top-owner-uid: the control-plane "workloadId" annotation, else the UID.
func topOwnerWorkloadId(runner *dromosv1alpha1.WorkloadRunner) string {
	if workloadId := runner.GetAnnotations()[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload]; workloadId != "" {
		return workloadId
	}
	return string(runner.UID)
}

func hasAnyKaiTopologyAnnotation(resource *unstructured.Unstructured) bool {
	annotations := resource.GetAnnotations()
	for _, key := range kaiTopologyAnnotationKeys {
		if _, ok := annotations[key]; ok {
			return true
		}
	}
	return false
}

func gpuNetworkAccelerationLabelKeys(runner *dromosv1alpha1.WorkloadRunner) []string {
	instr := runner.Spec.Instructions
	if instr == nil || instr.GpuOptimizations == nil || instr.GpuOptimizations.Mnnvl == nil {
		return nil
	}
	return instr.GpuOptimizations.Mnnvl.GpuNetworkAccelerationLabelKeys
}
