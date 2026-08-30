package runaimetrics

type RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15ResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15 struct {
	Filters        RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15Filters
	ResponseLabels RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15ResponseLabels
	Quantile       *float32
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) GetName() string {
	return "runai_avg_allocated_gpu_count_per_project_nodepool_top15"
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgAllocatedGpuCountPerProjectNodePoolTop15) IsQuantile() bool {
	return false
}
