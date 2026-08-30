package runaimetrics

type RunaiDeservedGpusCountPerQueuesFilters struct {
	ClusterId      []string `mapstructure:"clusterId,omitempty"`
	Project        []string `mapstructure:"project,omitempty"`
	NodePool       []string `mapstructure:"nodepool,omitempty"`
	Department     []string `mapstructure:"department,omitempty"`
	DepartmentName []string `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedGpusCountPerQueuesResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	Project        bool `mapstructure:"project,omitempty"`
	Nodepool       bool `mapstructure:"nodepool,omitempty"`
	Department     bool `mapstructure:"department,omitempty"`
	DepartmentName bool `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedGpusCountPerQueuesMetric struct {
	Filters        RunaiDeservedGpusCountPerQueuesFilters
	ResponseLabels RunaiDeservedGpusCountPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiDeservedGpusCountPerQueuesMetric) GetName() string {
	return "runai_deserved_gpu_count_per_queues"
}

func (m RunaiDeservedGpusCountPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDeservedGpusCountPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDeservedGpusCountPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDeservedGpusCountPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiDeservedGpusCountPerQueuesMetric) IsQuantile() bool {
	return false
}
