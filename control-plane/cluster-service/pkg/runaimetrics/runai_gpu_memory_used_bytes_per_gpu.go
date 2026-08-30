package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUsedBytesPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerGpuMetric struct {
	Filters        RunaiGpuMemoryUsedBytesPerGpuFilters
	ResponseLabels RunaiGpuMemoryUsedBytesPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) GetName() string {
	return "runai_gpu_memory_used_bytes_per_gpu"
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUsedBytesPerGpuMetric) IsQuantile() bool {
	return false
}
