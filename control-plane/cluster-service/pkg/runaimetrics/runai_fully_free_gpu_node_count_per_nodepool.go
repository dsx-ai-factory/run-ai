package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFullyFreeGpuNodeCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerNodepoolMetric struct {
	Filters        RunaiFullyFreeGpuNodeCountPerNodepoolFilters
	ResponseLabels RunaiFullyFreeGpuNodeCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) GetName() string {
	return "runai_fully_free_gpu_node_count_per_nodepool"
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiFullyFreeGpuNodeCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
