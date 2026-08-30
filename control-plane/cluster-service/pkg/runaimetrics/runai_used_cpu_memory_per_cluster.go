package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesPerClusterMetric struct {
	Filters        RunaiUsedMemoryBytesPerClusterFilters
	ResponseLabels RunaiUsedMemoryBytesPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesPerClusterMetric) GetName() string {
	return "runai_used_memory_bytes_per_cluster"
}

func (m RunaiUsedMemoryBytesPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedMemoryBytesPerClusterMetric) IsQuantile() bool {
	return false
}
