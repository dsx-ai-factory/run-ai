package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuLimitCoresPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuLimitCoresPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiCpuLimitCoresPerWorkloadMetric struct {
	Filters        RunaiCpuLimitCoresPerWorkloadFilters
	ResponseLabels RunaiCpuLimitCoresPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) GetName() string {
	return "runai_cpu_limit_cores_per_workload"
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuLimitCoresPerWorkloadMetric) IsQuantile() bool {
	return false
}
