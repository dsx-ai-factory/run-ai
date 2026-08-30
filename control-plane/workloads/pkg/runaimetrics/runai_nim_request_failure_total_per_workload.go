package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimRequestFailureTotalPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimRequestFailureTotalPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimRequestFailureTotalPerWorkloadMetric struct {
	Filters        RunaiNimRequestFailureTotalPerWorkloadFilters
	ResponseLabels RunaiNimRequestFailureTotalPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) GetName() string {
	return "runai_nim_request_failure_total_per_workload"
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimRequestFailureTotalPerWorkloadMetric) IsQuantile() bool {
	return false
}
