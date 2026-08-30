package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleGpuHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiIdleGpuHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiIdleGpuHoursPerQueueMetric struct {
	Filters        RunaiIdleGpuHoursPerQueueFilters
	ResponseLabels RunaiIdleGpuHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiIdleGpuHoursPerQueueMetric) GetName() string {
	return "runai_idle_gpu_hours_per_queue"
}

func (m RunaiIdleGpuHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleGpuHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleGpuHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleGpuHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiIdleGpuHoursPerQueueMetric) IsQuantile() bool {
	return false
}
