package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiTotalMemoryBytesPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiTotalMemoryBytesPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiTotalMemoryBytesPerClusterMetric struct {
	Filters        RunaiTotalMemoryBytesPerClusterFilters
	ResponseLabels RunaiTotalMemoryBytesPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiTotalMemoryBytesPerClusterMetric) GetName() string {
	return "runai_total_memory_bytes_per_cluster"
}

func (m RunaiTotalMemoryBytesPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiTotalMemoryBytesPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiTotalMemoryBytesPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiTotalMemoryBytesPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiTotalMemoryBytesPerClusterMetric) IsQuantile() bool {
	return false
}
