/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/google/uuid"
)

func GetIdleAllocatedGpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiIdleAllocatedGpuCountPerNodeMetric{
			Filters: runaimetrics.RunaiIdleAllocatedGpuCountPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiIdleAllocatedGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiIdleAllocatedGpuCountPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiIdleAllocatedGpuCountPerClusterMetric{
			Filters: runaimetrics.RunaiIdleAllocatedGpuCountPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiIdleAllocatedGpuCountPerTenantMetric{
			Filters: runaimetrics.RunaiIdleAllocatedGpuCountPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
