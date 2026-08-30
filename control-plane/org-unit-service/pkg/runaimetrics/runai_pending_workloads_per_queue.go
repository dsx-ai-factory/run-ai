package runaimetrics

type RunaiPendingWorkloadsPerQueueFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiPendingWorkloadsPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiPendingWorkloadsPerQueueMetric struct {
	Filters        RunaiPendingWorkloadsPerQueueFilters
	ResponseLabels RunaiPendingWorkloadsPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiPendingWorkloadsPerQueueMetric) GetName() string {
	return "runai_pending_workloads_per_queue"
}

func (m RunaiPendingWorkloadsPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiPendingWorkloadsPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiPendingWorkloadsPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiPendingWorkloadsPerQueueMetric) IsAggregation() bool {
	return false
}

func (m RunaiPendingWorkloadsPerQueueMetric) IsQuantile() bool {
	return false
}
