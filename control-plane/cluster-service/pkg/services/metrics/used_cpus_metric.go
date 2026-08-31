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

func GetUsedCpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiUsedCpuCoresPerNodeMetric{
			Filters: runaimetrics.RunaiUsedCpuCoresPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiUsedCpuCoresPerNodepoolMetric{
			Filters: runaimetrics.RunaiUsedCpuCoresPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiUsedCpuCoresPerClusterMetric{
			Filters: runaimetrics.RunaiUsedCpuCoresPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiUsedCpuCoresPerTenantMetric{
			Filters: runaimetrics.RunaiUsedCpuCoresPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
