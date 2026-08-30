package runaimetrics

type RunaiGpuMemoryUtilizationPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiGpuMemoryUtilizationPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiGpuMemoryUtilizationPerQueuesMetric struct {
	Filters        RunaiGpuMemoryUtilizationPerQueuesFilters
	ResponseLabels RunaiGpuMemoryUtilizationPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) GetName() string {
	return "runai_gpu_memory_utilization_per_queues"
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryUtilizationPerQueuesMetric) IsQuantile() bool {
	return false
}
