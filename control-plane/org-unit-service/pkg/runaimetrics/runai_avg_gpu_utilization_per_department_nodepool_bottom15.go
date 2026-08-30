package runaimetrics

type RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15Filters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
}

type RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15Labels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	NodePool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15 struct {
	Filters        RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15Filters `mapstructure:"filters,omitempty"`
	ResponseLabels RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15Labels  `mapstructure:"labels,omitempty"`
	Quantile       *float32
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) GetName() string {
	return "runai_avg_gpu_utilization_per_department_nodepool_bottom15"
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) IsAggregation() bool {
	return true
}

func (m RunaiAvgGpuUtilizationPerDepartmentNodePoolBottom15) IsQuantile() bool {
	return false
}
