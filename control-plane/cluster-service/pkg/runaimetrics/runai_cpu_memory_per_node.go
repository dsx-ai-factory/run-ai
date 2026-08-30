package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiTotalMemoryBytesPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiTotalMemoryBytesPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiTotalMemoryBytesPerNodeMetric struct {
	Filters        RunaiTotalMemoryBytesPerNodeFilters
	ResponseLabels RunaiTotalMemoryBytesPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiTotalMemoryBytesPerNodeMetric) GetName() string {
	return "runai_total_memory_bytes_per_node"
}

func (m RunaiTotalMemoryBytesPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiTotalMemoryBytesPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiTotalMemoryBytesPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiTotalMemoryBytesPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiTotalMemoryBytesPerNodeMetric) IsQuantile() bool {
	return false
}
