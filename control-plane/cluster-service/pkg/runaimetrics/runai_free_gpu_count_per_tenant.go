package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFreeGpuCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiFreeGpuCountPerTenantResponseLabels struct{}

type RunaiFreeGpuCountPerTenantMetric struct {
	Filters        RunaiFreeGpuCountPerTenantFilters
	ResponseLabels RunaiFreeGpuCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiFreeGpuCountPerTenantMetric) GetName() string {
	return "runai_free_gpu_count_per_tenant"
}

func (m RunaiFreeGpuCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFreeGpuCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFreeGpuCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFreeGpuCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiFreeGpuCountPerTenantMetric) IsQuantile() bool {
	return false
}
