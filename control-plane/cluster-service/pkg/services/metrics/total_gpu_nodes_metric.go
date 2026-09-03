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

func GetTotalGpuNodesMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiGpuNodesCountPerNodeMetric{
			Filters: runaimetrics.RunaiGpuNodesCountPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiGpuNodesCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuNodesCountPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiGpuNodesCountPerClusterMetric{
			Filters: runaimetrics.RunaiGpuNodesCountPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiGpuNodesCountPerTenantMetric{
			Filters: runaimetrics.RunaiGpuNodesCountPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
