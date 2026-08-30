package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func GetUsedCpuMemoryMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiUsedMemoryBytesPerNodeMetric{
			Filters: runaimetrics.RunaiUsedMemoryBytesPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiUsedMemoryBytesPerNodepoolMetric{
			Filters: runaimetrics.RunaiUsedMemoryBytesPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiUsedMemoryBytesPerClusterMetric{
			Filters: runaimetrics.RunaiUsedMemoryBytesPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiUsedMemoryBytesPerTenantMetric{
			Filters: runaimetrics.RunaiUsedMemoryBytesPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
