package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFreeGpuCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiFreeGpuCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiFreeGpuCountPerClusterMetric struct {
	Filters        RunaiFreeGpuCountPerClusterFilters
	ResponseLabels RunaiFreeGpuCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiFreeGpuCountPerClusterMetric) GetName() string {
	return "runai_free_gpu_count_per_cluster"
}

func (m RunaiFreeGpuCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFreeGpuCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFreeGpuCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFreeGpuCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiFreeGpuCountPerClusterMetric) IsQuantile() bool {
	return false
}
