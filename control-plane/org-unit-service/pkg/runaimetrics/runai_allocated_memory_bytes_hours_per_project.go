package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerProjectMetric struct {
	Filters        RunaiAllocatedMemoryBytesHoursPerProjectFilters
	ResponseLabels RunaiAllocatedMemoryBytesHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) GetName() string {
	return "runai_allocated_memory_bytes_hours_per_project"
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedMemoryBytesHoursPerProjectMetric) IsQuantile() bool {
	return false
}
