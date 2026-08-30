package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiDepartmentsGpuQuotaPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiDepartmentsGpuQuotaPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiDepartmentsGpuQuotaPerClusterMetric struct {
	Filters        RunaiDepartmentsGpuQuotaPerClusterFilters
	ResponseLabels RunaiDepartmentsGpuQuotaPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) GetName() string {
	return "runai_departments_gpu_quota_per_cluster"
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiDepartmentsGpuQuotaPerClusterMetric) IsQuantile() bool {
	return false
}
