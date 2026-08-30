package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerClusterMetric struct {
	Filters        RunaiUsedMemoryBytesHoursPerClusterFilters
	ResponseLabels RunaiUsedMemoryBytesHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) GetName() string {
	return "runai_used_memory_bytes_hours_per_cluster"
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedMemoryBytesHoursPerClusterMetric) IsQuantile() bool {
	return false
}
