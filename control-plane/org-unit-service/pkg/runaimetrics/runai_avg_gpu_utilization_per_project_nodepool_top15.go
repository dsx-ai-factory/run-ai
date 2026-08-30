package runaimetrics

type RunaiAvgGpuUtilizationPerProjectNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuUtilizationPerProjectNodePoolTop15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuUtilizationPerProjectNodePoolTop15 struct {
	Filters        RunaiAvgGpuUtilizationPerProjectNodePoolTop15Filters
	ResponseLabels RunaiAvgGpuUtilizationPerProjectNodePoolTop15Labels
	Quantile       *float32
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) GetName() string {
	return "runai_avg_gpu_utilization_per_project_nodepool_top15"
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuUtilizationPerProjectNodePoolTop15) IsQuantile() bool {
	return false
}
