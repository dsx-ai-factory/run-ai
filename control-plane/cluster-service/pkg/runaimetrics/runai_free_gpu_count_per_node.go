package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFreeGpuCountPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiFreeGpuCountPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiFreeGpuCountPerNodeMetric struct {
	Filters        RunaiFreeGpuCountPerNodeFilters
	ResponseLabels RunaiFreeGpuCountPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiFreeGpuCountPerNodeMetric) GetName() string {
	return "runai_free_gpu_count_per_node"
}

func (m RunaiFreeGpuCountPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFreeGpuCountPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFreeGpuCountPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFreeGpuCountPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiFreeGpuCountPerNodeMetric) IsQuantile() bool {
	return false
}
