package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryUtilizationPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiMemoryUtilizationPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiMemoryUtilizationPerClusterMetric struct {
	Filters        RunaiMemoryUtilizationPerClusterFilters
	ResponseLabels RunaiMemoryUtilizationPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUtilizationPerClusterMetric) GetName() string {
	return "runai_memory_utilization_per_cluster"
}

func (m RunaiMemoryUtilizationPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUtilizationPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUtilizationPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUtilizationPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUtilizationPerClusterMetric) IsQuantile() bool {
	return false
}
