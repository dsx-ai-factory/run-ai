package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiUsedCpuCoresPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiUsedCpuCoresPerNodeMetric struct {
	Filters        RunaiUsedCpuCoresPerNodeFilters
	ResponseLabels RunaiUsedCpuCoresPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresPerNodeMetric) GetName() string {
	return "runai_used_cpu_cores_per_node"
}

func (m RunaiUsedCpuCoresPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedCpuCoresPerNodeMetric) IsQuantile() bool {
	return false
}
