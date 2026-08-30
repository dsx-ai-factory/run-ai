package runaimetrics

type RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15 struct {
	Filters        RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) GetName() string {
	return "runai_avg_gpu_utilization_per_department_nodepool_top15"
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolTop15) IsQuantile() bool {
	return false
}
