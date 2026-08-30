package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUsedBytesPerPodPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerPodPerGpuMetric struct {
	Filters        RunaiGpuMemoryUsedBytesPerPodPerGpuFilters
	ResponseLabels RunaiGpuMemoryUsedBytesPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_memory_used_bytes_per_pod_per_gpu"
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUsedBytesPerPodPerGpuMetric) IsQuantile() bool {
	return false
}
