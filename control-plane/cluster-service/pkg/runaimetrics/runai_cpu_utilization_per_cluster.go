package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuUtilizationPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuUtilizationPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuUtilizationPerClusterMetric struct {
	Filters        RunaiCpuUtilizationPerClusterFilters
	ResponseLabels RunaiCpuUtilizationPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUtilizationPerClusterMetric) GetName() string {
	return "runai_cpu_utilization_per_cluster"
}

func (m RunaiCpuUtilizationPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUtilizationPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUtilizationPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUtilizationPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUtilizationPerClusterMetric) IsQuantile() bool {
	return false
}
