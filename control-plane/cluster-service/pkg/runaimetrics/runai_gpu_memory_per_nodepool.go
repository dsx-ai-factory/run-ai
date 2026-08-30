package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryBytesPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuMemoryBytesPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuMemoryBytesPerNodepoolMetric struct {
	Filters        RunaiGpuMemoryBytesPerNodepoolFilters
	ResponseLabels RunaiGpuMemoryBytesPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) GetName() string {
	return "runai_total_gpu_memory_byte_per_nodepool"
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryBytesPerNodepoolMetric) IsQuantile() bool {
	return false
}
