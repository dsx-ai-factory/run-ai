package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiAllocatedMemoryBytesPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiAllocatedMemoryBytesPerNodeMetric struct {
	Filters        RunaiAllocatedMemoryBytesPerNodeFilters
	ResponseLabels RunaiAllocatedMemoryBytesPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) GetName() string {
	return "runai_allocated_memory_bytes_per_node"
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryBytesPerNodeMetric) IsQuantile() bool {
	return false
}
