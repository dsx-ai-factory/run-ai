package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuUtilizationPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiGpuUtilizationPerWorkloadMetric struct {
	Filters        RunaiGpuUtilizationPerWorkloadFilters
	ResponseLabels RunaiGpuUtilizationPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerWorkloadMetric) GetName() string {
	return "runai_gpu_utilization_per_workload"
}

func (m RunaiGpuUtilizationPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerWorkloadMetric) IsQuantile() bool {
	return false
}
