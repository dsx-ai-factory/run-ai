package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
	Le         bool `mapstructure:"le,omitempty"`
}

type RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric struct {
	Filters        RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadFilters
	ResponseLabels RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) GetName() string {
	return "runai_nim_time_to_first_token_seconds_bucket_per_workload"
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric) IsQuantile() bool {
	return false
}
