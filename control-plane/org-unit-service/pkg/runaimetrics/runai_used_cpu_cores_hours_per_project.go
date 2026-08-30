package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiUsedCpuCoresHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiUsedCpuCoresHoursPerProjectMetric struct {
	Filters        RunaiUsedCpuCoresHoursPerProjectFilters
	ResponseLabels RunaiUsedCpuCoresHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) GetName() string {
	return "runai_used_cpu_cores_hours_per_project"
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedCpuCoresHoursPerProjectMetric) IsQuantile() bool {
	return false
}
