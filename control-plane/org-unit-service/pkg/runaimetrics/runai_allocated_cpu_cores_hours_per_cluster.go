package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerClusterMetric struct {
	Filters        RunaiAllocatedCpuCoresHoursPerClusterFilters
	ResponseLabels RunaiAllocatedCpuCoresHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) GetName() string {
	return "runai_allocated_cpu_cores_hours_per_cluster"
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedCpuCoresHoursPerClusterMetric) IsQuantile() bool {
	return false
}
