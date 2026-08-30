package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleGpuHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiIdleGpuHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiIdleGpuHoursPerDepartmentMetric struct {
	Filters        RunaiIdleGpuHoursPerDepartmentFilters
	ResponseLabels RunaiIdleGpuHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) GetName() string {
	return "runai_idle_gpu_hours_per_department"
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiIdleGpuHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
