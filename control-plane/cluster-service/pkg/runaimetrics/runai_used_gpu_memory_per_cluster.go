package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedGpuMemoryBytesPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerClusterMetric struct {
	Filters        RunaiUsedGpuMemoryBytesPerClusterFilters
	ResponseLabels RunaiUsedGpuMemoryBytesPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) GetName() string {
	return "runai_used_gpu_memory_bytes_per_cluster"
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedGpuMemoryBytesPerClusterMetric) IsQuantile() bool {
	return false
}
