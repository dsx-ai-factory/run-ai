package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryUsedBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryUsedBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryUsedBytesPerWorkloadMetric struct {
	Filters        RunaiMemoryUsedBytesPerWorkloadFilters
	ResponseLabels RunaiMemoryUsedBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) GetName() string {
	return "runai_memory_used_bytes_per_workload"
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUsedBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
