package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAvgWorkloadWaitTimeSecondsPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAvgWorkloadWaitTimeSecondsPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric struct {
	Filters        RunaiAvgWorkloadWaitTimeSecondsPerClusterFilters
	ResponseLabels RunaiAvgWorkloadWaitTimeSecondsPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) GetName() string {
	return "runai_avg_workload_wait_time_seconds_per_cluster"
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric) IsQuantile() bool {
	return false
}
