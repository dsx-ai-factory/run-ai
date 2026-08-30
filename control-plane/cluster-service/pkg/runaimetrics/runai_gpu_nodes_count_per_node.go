package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuNodesCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuNodesCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiGpuNodesCountPerNodeMetric struct {
	Filters        RunaiGpuNodesCountPerNodeFilters
	ResponseLabels RunaiGpuNodesCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNodesCountPerNodeMetric) GetName() string {
	return "runai_gpu_nodes_count_per_node"
}

func (m RunaiGpuNodesCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNodesCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNodesCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNodesCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNodesCountPerNodeMetric) IsQuantile() bool {
	return false
}
