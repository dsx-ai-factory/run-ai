package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryUsedBytesPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiMemoryUsedBytesPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiMemoryUsedBytesPerPodMetric struct {
	Filters        RunaiMemoryUsedBytesPerPodFilters
	ResponseLabels RunaiMemoryUsedBytesPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUsedBytesPerPodMetric) GetName() string {
	return "runai_memory_used_bytes_per_pod"
}

func (m RunaiMemoryUsedBytesPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUsedBytesPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUsedBytesPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUsedBytesPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUsedBytesPerPodMetric) IsQuantile() bool {
	return false
}
