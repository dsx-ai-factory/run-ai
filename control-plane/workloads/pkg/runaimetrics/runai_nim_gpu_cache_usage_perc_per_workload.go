package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimGpuCacheUsagePercPerWorkloadFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	WorkloadId []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiNimGpuCacheUsagePercPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiNimGpuCacheUsagePercPerWorkloadMetric struct {
	Filters        RunaiNimGpuCacheUsagePercPerWorkloadFilters
	ResponseLabels RunaiNimGpuCacheUsagePercPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) GetName() string {
	return "runai_nim_gpu_cache_usage_perc_per_workload"
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimGpuCacheUsagePercPerWorkloadMetric) IsQuantile() bool {
	return false
}
