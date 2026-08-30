package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiUsedMemoryBytesPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiUsedMemoryBytesPerNodeMetric struct {
	Filters        RunaiUsedMemoryBytesPerNodeFilters
	ResponseLabels RunaiUsedMemoryBytesPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesPerNodeMetric) GetName() string {
	return "runai_used_memory_bytes_per_node"
}

func (m RunaiUsedMemoryBytesPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedMemoryBytesPerNodeMetric) IsQuantile() bool {
	return false
}
