package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuUtilizationPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiGpuUtilizationPerNodeMetric struct {
	Filters        RunaiGpuUtilizationPerNodeFilters
	ResponseLabels RunaiGpuUtilizationPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerNodeMetric) GetName() string {
	return "runai_gpu_utilization_per_node"
}

func (m RunaiGpuUtilizationPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerNodeMetric) IsQuantile() bool {
	return false
}
