package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiDepartmentsGpuQuotaPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiDepartmentsGpuQuotaPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiDepartmentsGpuQuotaPerNodepoolMetric struct {
	Filters        RunaiDepartmentsGpuQuotaPerNodepoolFilters
	ResponseLabels RunaiDepartmentsGpuQuotaPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) GetName() string {
	return "runai_departments_gpu_quota_per_nodepool"
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiDepartmentsGpuQuotaPerNodepoolMetric) IsQuantile() bool {
	return false
}
