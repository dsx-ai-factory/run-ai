package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedGpuCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedGpuCountPerNodepoolMetric struct {
	Filters        RunaiAllocatedGpuCountPerNodepoolFilters
	ResponseLabels RunaiAllocatedGpuCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) GetName() string {
	return "runai_allocated_gpu_count_per_nodepool"
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
