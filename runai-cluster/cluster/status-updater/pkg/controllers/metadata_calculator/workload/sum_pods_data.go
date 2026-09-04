/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	podcalculator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/pod"
	v1 "k8s.io/api/core/v1"
)

func CalcPodsUsedNodesMetadata(workloadMetadata *metadata.WorkloadMetadata, pods []v1.Pod) {
	var usedNodesList []string
	usedNodesMap := make(map[string]bool)
	for _, item := range pods {
		pod := item
		podExtendedPhase := podcalculator.GetExtendedPhase(&pod)

		if podcalculator.IsPodBound(podExtendedPhase) && !usedNodesMap[pod.Spec.NodeName] {
			usedNodesMap[pod.Spec.NodeName] = true
			usedNodesList = append(usedNodesList, pod.Spec.NodeName)
		}
	}

	if len(usedNodesList) > 0 {
		workloadMetadata.UsedNodes = strings.Join(usedNodesList, ", ")
	}
}

func SumPodGpuMemoryResources(workloadMetadata *metadata.WorkloadMetadata, pods []v1.Pod) {
	for _, item := range pods {
		pod := item

		workloadMetadata.RequestedGPUsMemory += podcalculator.GetGpuMemoryRequest(&pod)
		workloadMetadata.CurrentRequestedGPUsMemory += podcalculator.GetGpuMemoryRequest(&pod)

		if podcalculator.IsPodBound(podcalculator.GetExtendedPhase(&pod)) {
			workloadMetadata.AllocatedGPUsMemory += podcalculator.GetGpuMemoryRequest(&pod)
		}
	}
}
