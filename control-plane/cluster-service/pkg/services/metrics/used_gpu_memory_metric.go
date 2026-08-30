package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func GetUsedGpuMemoryMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiUsedGpuMemoryBytesPerNodeMetric{
			Filters: runaimetrics.RunaiUsedGpuMemoryBytesPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiUsedGpuMemoryBytesPerNodepoolMetric{
			Filters: runaimetrics.RunaiUsedGpuMemoryBytesPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiUsedGpuMemoryBytesPerClusterMetric{
			Filters: runaimetrics.RunaiUsedGpuMemoryBytesPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiUsedGpuMemoryBytesPerTenantMetric{
			Filters: runaimetrics.RunaiUsedGpuMemoryBytesPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
