package runaimetrics

type RunaiMemoryUtilizationPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiMemoryUtilizationPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiMemoryUtilizationPerQueuesMetric struct {
	Filters        RunaiMemoryUtilizationPerQueuesFilters
	ResponseLabels RunaiMemoryUtilizationPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUtilizationPerQueuesMetric) GetName() string {
	return "runai_memory_utilization_per_queues"
}

func (m RunaiMemoryUtilizationPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUtilizationPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUtilizationPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUtilizationPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUtilizationPerQueuesMetric) IsQuantile() bool {
	return false
}
