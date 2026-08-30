package runaimetrics

type RunaiAllocatedCpuPerQueueFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedCpuPerQueueMetric struct {
	Filters        RunaiAllocatedCpuPerQueueFilters
	ResponseLabels RunaiAllocatedCpuPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuPerQueueMetric) GetName() string {
	return "runai_allocated_cpu_per_queue"
}

func (m RunaiAllocatedCpuPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuPerQueueMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuPerQueueMetric) IsQuantile() bool {
	return false
}
