package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
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
