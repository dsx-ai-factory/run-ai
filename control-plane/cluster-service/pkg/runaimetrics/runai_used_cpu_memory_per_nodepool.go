package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedMemoryBytesPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedMemoryBytesPerNodepoolMetric struct {
	Filters        RunaiUsedMemoryBytesPerNodepoolFilters
	ResponseLabels RunaiUsedMemoryBytesPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) GetName() string {
	return "runai_used_memory_bytes_per_nodepool"
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedMemoryBytesPerNodepoolMetric) IsQuantile() bool {
	return false
}
