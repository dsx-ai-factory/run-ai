package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestsRunningPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestsRunningPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestsRunningPerPodMetric struct {
	Filters        RunaiNimNumRequestsRunningPerPodFilters
	ResponseLabels RunaiNimNumRequestsRunningPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestsRunningPerPodMetric) GetName() string {
	return "runai_nim_num_requests_running_per_pod"
}

func (m RunaiNimNumRequestsRunningPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestsRunningPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestsRunningPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestsRunningPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestsRunningPerPodMetric) IsQuantile() bool {
	return false
}
