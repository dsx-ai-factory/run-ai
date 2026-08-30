package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerProjectMetric struct {
	Filters        RunaiAllocatedCpuCoresHoursPerProjectFilters
	ResponseLabels RunaiAllocatedCpuCoresHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) GetName() string {
	return "runai_allocated_cpu_cores_hours_per_project"
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedCpuCoresHoursPerProjectMetric) IsQuantile() bool {
	return false
}
