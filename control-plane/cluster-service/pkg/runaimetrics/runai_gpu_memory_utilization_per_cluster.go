package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryUtilizationPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryUtilizationPerClusterMetric struct {
	Filters        RunaiGpuMemoryUtilizationPerClusterFilters
	ResponseLabels RunaiGpuMemoryUtilizationPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) GetName() string {
	return "runai_gpu_memory_utilization_per_cluster"
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUtilizationPerClusterMetric) IsQuantile() bool {
	return false
}
