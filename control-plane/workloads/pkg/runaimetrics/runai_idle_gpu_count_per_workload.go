package runaimetrics

type RunaiIdleGpuCountPerWorkloadFilters struct {
	WorkloadId []string `mapstructure:"workload_id,omitempty"`
}

type RunaiIdleGpuCountPerWorkloadResponseLabels struct {
	WorkloadId bool `mapstructure:"workload_id,omitempty"`
}

type RunaiIdleGpuCountPerWorkloadMetric struct {
	Filters        RunaiIdleGpuCountPerWorkloadFilters
	ResponseLabels RunaiIdleGpuCountPerWorkloadResponseLabels
	Quantile       *float32
}

func (m RunaiIdleGpuCountPerWorkloadMetric) GetName() string {
	return "runai_idle_gpu_count_per_workload"
}

func (m RunaiIdleGpuCountPerWorkloadMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleGpuCountPerWorkloadMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleGpuCountPerWorkloadMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleGpuCountPerWorkloadMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleGpuCountPerWorkloadMetric) IsQuantile() bool {
	return false
}
