package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuCountPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuCountPerTenantMetric struct {
	Filters        RunaiGpuCountPerTenantFilters
	ResponseLabels RunaiGpuCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiGpuCountPerTenantMetric) GetName() string {
	return "runai_gpu_count_per_tenant"
}

func (m RunaiGpuCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuCountPerTenantMetric) IsQuantile() bool {
	return false
}
