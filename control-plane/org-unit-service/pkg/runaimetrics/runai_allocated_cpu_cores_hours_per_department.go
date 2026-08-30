package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerDepartmentMetric struct {
	Filters        RunaiAllocatedCpuCoresHoursPerDepartmentFilters
	ResponseLabels RunaiAllocatedCpuCoresHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) GetName() string {
	return "runai_allocated_cpu_cores_hours_per_department"
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedCpuCoresHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
