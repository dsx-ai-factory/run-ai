package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUtilizationAvgPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiGpuMemoryUtilizationAvgPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiGpuMemoryUtilizationAvgPerProjectMetric struct {
	Filters        RunaiGpuMemoryUtilizationAvgPerProjectFilters
	ResponseLabels RunaiGpuMemoryUtilizationAvgPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) GetName() string {
	return "runai_gpu_memory_utilization_avg_per_project"
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiGpuMemoryUtilizationAvgPerProjectMetric) IsQuantile() bool {
	return false
}
