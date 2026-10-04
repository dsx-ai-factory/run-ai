/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
)

const (
	PgUsedNodes                  = "runai-used-nodes"
	PgRequestedGPUs              = "runai-podgroup-requested-gpus"
	PgRequestedGPUsMemory        = "runai-podgroup-requested-gpus-memory"
	PgTotalRequestedGPUs         = "runai-total-requested-gpus"
	PgTotalRequestedGPUsMemory   = "runai-total-requested-gpus-memory"
	PgCurrentRequestedGPUs       = "runai-current-requested-gpus"
	PgCurrentRequestedGPUsMemory = "runai-current-requested-gpus-memory"
	PgAllocatedGPUs              = "runai-current-allocated-gpus"
	PgAllocatedGPUsMemory        = "runai-current-allocated-gpus-memory"
)

func SumOwnedPodGroupResources(workloadMetadata *metadata.WorkloadMetadata, ownedPodGroups []kaiv2alpha2.PodGroup) {
	for _, podGroup := range ownedPodGroups {
		if resource, exists := podGroup.Status.ResourcesStatus.Requested[clustercommonconstants.NvidiaGpuResourceName]; exists {
			workloadMetadata.RequestedGPUs += resource.AsApproximateFloat64()
			workloadMetadata.CurrentRequestedGPUs += resource.AsApproximateFloat64()
		}
		if resource, exists := podGroup.Status.ResourcesStatus.Allocated[clustercommonconstants.NvidiaGpuResourceName]; exists {
			workloadMetadata.AllocatedGpus += resource.AsApproximateFloat64()
		}
	}
}
