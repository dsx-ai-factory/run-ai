package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedCpuCoresPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedCpuCoresPerNodepoolMetric struct {
	Filters        RunaiUsedCpuCoresPerNodepoolFilters
	ResponseLabels RunaiUsedCpuCoresPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) GetName() string {
	return "runai_used_cpu_cores_per_nodepool"
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedCpuCoresPerNodepoolMetric) IsQuantile() bool {
	return false
}
