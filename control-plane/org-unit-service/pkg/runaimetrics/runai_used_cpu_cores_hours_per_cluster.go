package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresHoursPerClusterMetric struct {
	Filters        RunaiUsedCpuCoresHoursPerClusterFilters
	ResponseLabels RunaiUsedCpuCoresHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) GetName() string {
	return "runai_used_cpu_cores_hours_per_cluster"
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedCpuCoresHoursPerClusterMetric) IsQuantile() bool {
	return false
}
