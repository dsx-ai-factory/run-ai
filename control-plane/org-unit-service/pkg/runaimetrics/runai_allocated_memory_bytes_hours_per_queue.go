package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerQueueMetric struct {
	Filters        RunaiAllocatedMemoryBytesHoursPerQueueFilters
	ResponseLabels RunaiAllocatedMemoryBytesHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) GetName() string {
	return "runai_allocated_memory_bytes_hours_per_queue"
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedMemoryBytesHoursPerQueueMetric) IsQuantile() bool {
	return false
}
