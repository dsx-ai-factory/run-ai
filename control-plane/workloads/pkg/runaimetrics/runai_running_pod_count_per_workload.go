package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiRunningPodCountPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
	JobId      []string    `mapstructure:"job_uuid,omitempty"`
}

type RunaiRunningPodCountPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
}

type RunaiRunningPodCountPerWorkloadMetric struct {
	Filters        RunaiRunningPodCountPerWorkloadFilters
	ResponseLabels RunaiRunningPodCountPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiRunningPodCountPerWorkloadMetric) GetName() string {
	return "runai_running_pod_count_per_workload"
}

func (m RunaiRunningPodCountPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiRunningPodCountPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiRunningPodCountPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiRunningPodCountPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiRunningPodCountPerWorkloadMetric) IsQuantile() bool {
	return false
}
