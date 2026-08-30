package runaimetrics

type RunaiPendingWorkloadsCountPerQueuesFilters struct {
	ClusterId      []string `mapstructure:"clusterId,omitempty"`
	Project        []string `mapstructure:"project,omitempty"`
	NodePool       []string `mapstructure:"nodepool,omitempty"`
	Department     []string `mapstructure:"department,omitempty"`
	DepartmentName []string `mapstructure:"department_name,omitempty"`
}

type RunaiPendingWorkloadsCountPerQueuesResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	Project        bool `mapstructure:"project,omitempty"`
	Nodepool       bool `mapstructure:"nodepool,omitempty"`
	Department     bool `mapstructure:"department,omitempty"`
	DepartmentName bool `mapstructure:"department_name,omitempty"`
}

type RunaiPendingWorkloadsCountPerQueuesMetric struct {
	Filters        RunaiPendingWorkloadsCountPerQueuesFilters
	ResponseLabels RunaiPendingWorkloadsCountPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) GetName() string {
	return "runai_pending_workloads_count_per_queues"
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiPendingWorkloadsCountPerQueuesMetric) IsQuantile() bool {
	return false
}
