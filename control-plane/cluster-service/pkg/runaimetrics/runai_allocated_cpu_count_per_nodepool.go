package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedCpuCoresPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedCpuCoresPerNodepoolMetric struct {
	Filters        RunaiAllocatedCpuCoresPerNodepoolFilters
	ResponseLabels RunaiAllocatedCpuCoresPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) GetName() string {
	return "runai_allocated_cpu_cores_per_nodepool"
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuCoresPerNodepoolMetric) IsQuantile() bool {
	return false
}
