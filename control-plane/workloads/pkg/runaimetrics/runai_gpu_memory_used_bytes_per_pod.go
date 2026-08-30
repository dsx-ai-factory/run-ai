package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUsedBytesPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerPodMetric struct {
	Filters        RunaiGpuMemoryUsedBytesPerPodFilters
	ResponseLabels RunaiGpuMemoryUsedBytesPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) GetName() string {
	return "runai_gpu_memory_used_bytes_per_pod"
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUsedBytesPerPodMetric) IsQuantile() bool {
	return false
}
