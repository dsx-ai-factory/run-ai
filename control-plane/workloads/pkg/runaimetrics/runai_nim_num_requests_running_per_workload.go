package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestsRunningPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestsRunningPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimNumRequestsRunningPerWorkloadMetric struct {
	Filters        RunaiNimNumRequestsRunningPerWorkloadFilters
	ResponseLabels RunaiNimNumRequestsRunningPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) GetName() string {
	return "runai_nim_num_requests_running_per_workload"
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestsRunningPerWorkloadMetric) IsQuantile() bool {
	return false
}
