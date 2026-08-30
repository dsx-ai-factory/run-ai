package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleAllocatedGpuCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerNodeMetric struct {
	Filters        RunaiIdleAllocatedGpuCountPerNodeFilters
	ResponseLabels RunaiIdleAllocatedGpuCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) GetName() string {
	return "runai_idle_allocated_gpu_count_per_node"
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleAllocatedGpuCountPerNodeMetric) IsQuantile() bool {
	return false
}
