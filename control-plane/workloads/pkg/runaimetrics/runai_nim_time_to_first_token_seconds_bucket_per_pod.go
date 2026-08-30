package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimTimeToFirstTokenSecondsBucketPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Le    bool `mapstructure:"le,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric struct {
	Filters        RunaiNimTimeToFirstTokenSecondsBucketPerPodFilters
	ResponseLabels RunaiNimTimeToFirstTokenSecondsBucketPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) GetName() string {
	return "runai_nim_time_to_first_token_seconds_bucket_per_pod"
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric) IsQuantile() bool {
	return false
}
