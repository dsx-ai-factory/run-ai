package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerWorkloadFilters struct {
	ClusterId      []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId     []string    `mapstructure:"workload_id,omitempty"`
	JobId          []string    `mapstructure:"job_uuid,omitempty"`
	JobName        []string    `mapstructure:"job_name,omitempty"`
	JobType        []string    `mapstructure:"job_type,omitempty"`
	Project        []string    `mapstructure:"project,omitempty"`
	Department     []string    `mapstructure:"department,omitempty"`
	WorkloadName   []string    `mapstructure:"workload_name,omitempty"`
	WorkloadType   []string    `mapstructure:"workload_type,omitempty"`
	User           []string    `mapstructure:"user,omitempty"`
	Nodepool       []string    `mapstructure:"nodepool,omitempty"`
	QueueName      []string    `mapstructure:"queue_name,omitempty"`
	IsGpu          []bool      `mapstructure:"is_gpu,omitempty"`
	DetailedStatus []string    `mapstructure:"detailed_status,omitempty"`
}

type RunaiAllocatedGpuCountPerWorkloadResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	JobId      bool `mapstructure:"job_uuid,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedGpuCountPerWorkloadMetric struct {
	Filters        RunaiAllocatedGpuCountPerWorkloadFilters
	ResponseLabels RunaiAllocatedGpuCountPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) GetName() string {
	return "runai_allocated_gpu_count_per_workload"
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerWorkloadMetric) IsQuantile() bool {
	return false
}
