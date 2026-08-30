package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiInferenceLatencyPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiInferenceLatencyPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiInferenceLatencyPerWorkloadMetric struct {
	Filters        RunaiInferenceLatencyPerWorkloadFilters
	ResponseLabels RunaiInferenceLatencyPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiInferenceLatencyPerWorkloadMetric) GetName() string {
	return "runai_inference_latency_per_workload"
}

func (m RunaiInferenceLatencyPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiInferenceLatencyPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiInferenceLatencyPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiInferenceLatencyPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiInferenceLatencyPerWorkloadMetric) IsQuantile() bool {
	return false
}
