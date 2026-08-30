package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFullyFreeGpuNodeCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerNodeMetric struct {
	Filters        RunaiFullyFreeGpuNodeCountPerNodeFilters
	ResponseLabels RunaiFullyFreeGpuNodeCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) GetName() string {
	return "runai_fully_free_gpu_node_count_per_node"
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiFullyFreeGpuNodeCountPerNodeMetric) IsQuantile() bool {
	return false
}
