package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryUsedBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuMemoryUsedBytesPerWorkloadMetric struct {
	Filters        RunaiGpuMemoryUsedBytesPerWorkloadFilters
	ResponseLabels RunaiGpuMemoryUsedBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) GetName() string {
	return "runai_gpu_memory_used_bytes_per_workload"
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUsedBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
