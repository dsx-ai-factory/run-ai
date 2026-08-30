package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryRequestedBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryRequestedBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryRequestedBytesPerWorkloadMetric struct {
	Filters        RunaiMemoryRequestedBytesPerWorkloadFilters
	ResponseLabels RunaiMemoryRequestedBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) GetName() string {
	return "runai_memory_requested_bytes_per_workload"
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryRequestedBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
