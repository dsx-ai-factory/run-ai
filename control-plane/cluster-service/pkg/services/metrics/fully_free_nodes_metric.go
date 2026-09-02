/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

func GetFullyFreeGpuNodesMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metricsClient.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeMetric{
			Filters: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiFullyFreeGpuNodeCountPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterMetric{
			Filters: runaimetrics.RunaiFullyFreeGpuNodeCountPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantMetric{
			Filters: runaimetrics.RunaiFullyFreeGpuNodeCountPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
