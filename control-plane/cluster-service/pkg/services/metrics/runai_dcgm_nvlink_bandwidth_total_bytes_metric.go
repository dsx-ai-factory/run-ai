package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
)

func GetDcgmNvlinkBandwidthTotalBytesMetric(clusterIds []uuid.UUID, metricGroupBy metricsClient.MetricGroupBy, nodeFilter []string) metricsClient.GroupedMetric {
	metric := runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesMetric{
		Filters: runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesFilters{
			ClusterId: clusterIds,
			Node:      nodeFilter,
		},
		GroupBy: metricGroupBy,
	}
	return metric
}
