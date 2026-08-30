package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuNodesCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuNodesCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuNodesCountPerNodepoolMetric struct {
	Filters        RunaiGpuNodesCountPerNodepoolFilters
	ResponseLabels RunaiGpuNodesCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNodesCountPerNodepoolMetric) GetName() string {
	return "runai_gpu_nodes_count_per_nodepool"
}

func (m RunaiGpuNodesCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNodesCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNodesCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNodesCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNodesCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
