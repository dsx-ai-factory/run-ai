package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryLimitBytesPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryLimitBytesPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiMemoryLimitBytesPerWorkloadMetric struct {
	Filters        RunaiMemoryLimitBytesPerWorkloadFilters
	ResponseLabels RunaiMemoryLimitBytesPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) GetName() string {
	return "runai_memory_limit_bytes_per_workload"
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryLimitBytesPerWorkloadMetric) IsQuantile() bool {
	return false
}
