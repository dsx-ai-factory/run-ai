package runaimetrics

type RunaiCpuUtilizationPerNodepoolsFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
	NodePool  []string `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuUtilizationPerNodepoolsResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuUtilizationPerNodepoolsMetric struct {
	Filters        RunaiCpuUtilizationPerNodepoolsFilters
	ResponseLabels RunaiCpuUtilizationPerNodepoolsResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) GetName() string {
	return "runai_cpu_utilization_per_nodepools"
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUtilizationPerNodepoolsMetric) IsQuantile() bool {
	return false
}
