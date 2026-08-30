package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiGpuCountPerNodeMetric struct {
	Filters        RunaiGpuCountPerNodeFilters
	ResponseLabels RunaiGpuCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiGpuCountPerNodeMetric) GetName() string {
	return "runai_gpu_count_per_node"
}

func (m RunaiGpuCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuCountPerNodeMetric) IsQuantile() bool {
	return false
}
