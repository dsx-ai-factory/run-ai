package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryBytesPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryBytesPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryBytesPerNodeMetric struct {
	Filters        RunaiGpuMemoryBytesPerNodeFilters
	ResponseLabels RunaiGpuMemoryBytesPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryBytesPerNodeMetric) GetName() string {
	return "runai_total_gpu_memory_byte_per_node"
}

func (m RunaiGpuMemoryBytesPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryBytesPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryBytesPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryBytesPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryBytesPerNodeMetric) IsQuantile() bool {
	return false
}
