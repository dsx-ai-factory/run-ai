package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedGpuMemoryBytesPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedGpuMemoryBytesPerTenantMetric struct {
	Filters        RunaiUsedGpuMemoryBytesPerTenantFilters
	ResponseLabels RunaiUsedGpuMemoryBytesPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) GetName() string {
	return "runai_used_gpu_memory_bytes_per_tenant"
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedGpuMemoryBytesPerTenantMetric) IsQuantile() bool {
	return false
}
