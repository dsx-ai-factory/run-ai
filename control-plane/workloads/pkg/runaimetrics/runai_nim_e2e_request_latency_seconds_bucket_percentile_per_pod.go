package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodResponseLabels struct {
	PodId         bool `mapstructure:"pod_uuid,omitempty"`
	QuantileLabel bool `mapstructure:"quantile_label,omitempty"`
}

type RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric struct {
	Filters        RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodFilters
	ResponseLabels RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) GetName() string {
	return "runai_nim_e2e_request_latency_seconds_bucket_percentile_per_pod"
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric) IsQuantile() bool {
	return false
}
