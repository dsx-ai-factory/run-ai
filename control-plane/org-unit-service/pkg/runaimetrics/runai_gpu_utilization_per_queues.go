package runaimetrics

type RunaiGpuUtilizationPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiGpuUtilizationPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiGpuUtilizationPerQueuesMetric struct {
	Filters        RunaiGpuUtilizationPerQueuesFilters
	ResponseLabels RunaiGpuUtilizationPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerQueuesMetric) GetName() string {
	return "runai_gpu_utilization_per_queues"
}

func (m RunaiGpuUtilizationPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerQueuesMetric) IsQuantile() bool {
	return false
}
