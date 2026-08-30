package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuCountPerNodepoolMetric struct {
	Filters        RunaiGpuCountPerNodepoolFilters
	ResponseLabels RunaiGpuCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuCountPerNodepoolMetric) GetName() string {
	return "runai_gpu_count_per_nodepool"
}

func (m RunaiGpuCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
