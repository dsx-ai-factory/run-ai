package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuUtilizationPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuUtilizationPerGpuMetric struct {
	Filters        RunaiGpuUtilizationPerGpuFilters
	ResponseLabels RunaiGpuUtilizationPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerGpuMetric) GetName() string {
	return "runai_gpu_utilization_per_gpu"
}

func (m RunaiGpuUtilizationPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerGpuMetric) IsQuantile() bool {
	return false
}
