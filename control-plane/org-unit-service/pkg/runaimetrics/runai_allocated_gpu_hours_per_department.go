package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedGpuHoursPerDepartmentMetric struct {
	Filters        RunaiAllocatedGpuHoursPerDepartmentFilters
	ResponseLabels RunaiAllocatedGpuHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) GetName() string {
	return "runai_allocated_gpu_hours_per_department"
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
