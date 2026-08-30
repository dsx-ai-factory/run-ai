package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedGpuHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedGpuHoursPerClusterMetric struct {
	Filters        RunaiAllocatedGpuHoursPerClusterFilters
	ResponseLabels RunaiAllocatedGpuHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) GetName() string {
	return "runai_allocated_gpu_hours_per_cluster"
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuHoursPerClusterMetric) IsQuantile() bool {
	return false
}
