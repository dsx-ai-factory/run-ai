package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiInferenceThroughputPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiInferenceThroughputPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiInferenceThroughputPerPodMetric struct {
	Filters        RunaiInferenceThroughputPerPodFilters
	ResponseLabels RunaiInferenceThroughputPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiInferenceThroughputPerPodMetric) GetName() string {
	return "runai_inference_throughput_per_pod"
}

func (m RunaiInferenceThroughputPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiInferenceThroughputPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiInferenceThroughputPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiInferenceThroughputPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiInferenceThroughputPerPodMetric) IsQuantile() bool {
	return false
}
