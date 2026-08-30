package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryUtilizationPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiMemoryUtilizationPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiMemoryUtilizationPerNodepoolMetric struct {
	Filters        RunaiMemoryUtilizationPerNodepoolFilters
	ResponseLabels RunaiMemoryUtilizationPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) GetName() string {
	return "runai_memory_utilization_per_nodepool"
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUtilizationPerNodepoolMetric) IsQuantile() bool {
	return false
}
