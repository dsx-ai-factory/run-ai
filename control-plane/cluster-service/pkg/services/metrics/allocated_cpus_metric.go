package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func GetAllocatedCpusMetric(clusterIdsFilter []uuid.UUID, nodePoolFilter []string, metricGroupBy MetricGroupBy) metrics_client.Metric {
	switch metricGroupBy {
	case GroupByNode:
		return runaimetrics.RunaiAllocatedCpuCoresPerNodeMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCoresPerNodeFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByNodePool:
		return runaimetrics.RunaiAllocatedCpuCoresPerNodepoolMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCoresPerNodepoolFilters{
				ClusterId: clusterIdsFilter,
				Nodepool:  nodePoolFilter,
			},
		}
	case GroupByCluster:
		return runaimetrics.RunaiAllocatedCpuCoresPerClusterMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCoresPerClusterFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	default: // GroupByTenant
		return runaimetrics.RunaiAllocatedCpuCoresPerTenantMetric{
			Filters: runaimetrics.RunaiAllocatedCpuCoresPerTenantFilters{
				ClusterId: clusterIdsFilter,
			},
		}
	}
}
