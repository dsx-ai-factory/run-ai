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

func GetTotalCpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiCpuCoresPerNodeMetric{
			Filters: runaimetrics.RunaiCpuCoresPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiCpuCoresPerNodepoolMetric{
			Filters: runaimetrics.RunaiCpuCoresPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiCpuCoresPerClusterMetric{
			Filters: runaimetrics.RunaiCpuCoresPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiCpuCoresPerTenantMetric{
			Filters: runaimetrics.RunaiCpuCoresPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
