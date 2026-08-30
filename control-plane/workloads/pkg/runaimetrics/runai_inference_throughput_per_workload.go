package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiInferenceThroughputPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiInferenceThroughputPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiInferenceThroughputPerWorkloadMetric struct {
	Filters        RunaiInferenceThroughputPerWorkloadFilters
	ResponseLabels RunaiInferenceThroughputPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiInferenceThroughputPerWorkloadMetric) GetName() string {
	return "runai_inference_throughput_per_workload"
}

func (m RunaiInferenceThroughputPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiInferenceThroughputPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiInferenceThroughputPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiInferenceThroughputPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiInferenceThroughputPerWorkloadMetric) IsQuantile() bool {
	return false
}
