package runaimetrics

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15 struct {
	Filters        RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) GetName() string {
	return "runai_avg_gpu_memory_utilization_per_department_nodepool_bottom15"
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuMemoryUtilizationPerDepartmentNodePoolBottom15) IsQuantile() bool {
	return false
}
