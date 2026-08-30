package runaimetrics

type RunaiGpuUtilizationPerNodepoolsFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
	NodePool  []string `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuUtilizationPerNodepoolsResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuUtilizationPerNodepoolsMetric struct {
	Filters        RunaiGpuUtilizationPerNodepoolsFilters
	ResponseLabels RunaiGpuUtilizationPerNodepoolsResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) GetName() string {
	return "runai_gpu_utilization_per_nodepools"
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerNodepoolsMetric) IsQuantile() bool {
	return false
}
