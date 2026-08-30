package runaimetrics

type RunaiCpuUtilizationPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiCpuUtilizationPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiCpuUtilizationPerQueuesMetric struct {
	Filters        RunaiCpuUtilizationPerQueuesFilters
	ResponseLabels RunaiCpuUtilizationPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUtilizationPerQueuesMetric) GetName() string {
	return "runai_cpu_utilization_per_queues"
}

func (m RunaiCpuUtilizationPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUtilizationPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUtilizationPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUtilizationPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUtilizationPerQueuesMetric) IsQuantile() bool {
	return false
}
