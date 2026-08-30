package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleGpuHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiIdleGpuHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiIdleGpuHoursPerProjectMetric struct {
	Filters        RunaiIdleGpuHoursPerProjectFilters
	ResponseLabels RunaiIdleGpuHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiIdleGpuHoursPerProjectMetric) GetName() string {
	return "runai_idle_gpu_hours_per_project"
}

func (m RunaiIdleGpuHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleGpuHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleGpuHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleGpuHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiIdleGpuHoursPerProjectMetric) IsQuantile() bool {
	return false
}
