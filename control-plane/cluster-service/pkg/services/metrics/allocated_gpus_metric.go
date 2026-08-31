/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func GetAllocatedGpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiAllocatedGpuCountPerNodeMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiAllocatedGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiAllocatedGpuCountPerClusterMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiAllocatedGpuCountPerTenantMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
