package runaimetrics

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15 struct {
	Filters        RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) GetName() string {
	return "runai_avg_allocated_gpu_count_per_department_nodepool_bottom15"
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgAllocatedGpuCountPerDepartmentNodePoolBottom15) IsQuantile() bool {
	return false
}
