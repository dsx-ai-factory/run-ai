package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuUtilizationPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiCpuUtilizationPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiCpuUtilizationPerNodeMetric struct {
	Filters        RunaiCpuUtilizationPerNodeFilters
	ResponseLabels RunaiCpuUtilizationPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUtilizationPerNodeMetric) GetName() string {
	return "runai_cpu_utilization_per_node"
}

func (m RunaiCpuUtilizationPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUtilizationPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUtilizationPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUtilizationPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUtilizationPerNodeMetric) IsQuantile() bool {
	return false
}
