package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationAvgPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiGpuUtilizationAvgPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiGpuUtilizationAvgPerProjectMetric struct {
	Filters        RunaiGpuUtilizationAvgPerProjectFilters
	ResponseLabels RunaiGpuUtilizationAvgPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) GetName() string {
	return "runai_gpu_utilization_avg_per_project"
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiGpuUtilizationAvgPerProjectMetric) IsQuantile() bool {
	return false
}
