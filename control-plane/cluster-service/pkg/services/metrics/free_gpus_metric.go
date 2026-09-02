/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

func GetFreeGpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiFreeGpuCountPerNodeMetric{
			Filters: runaimetrics.RunaiFreeGpuCountPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiFreeGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiFreeGpuCountPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiFreeGpuCountPerClusterMetric{
			Filters: runaimetrics.RunaiFreeGpuCountPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiFreeGpuCountPerTenantMetric{
			Filters: runaimetrics.RunaiFreeGpuCountPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
