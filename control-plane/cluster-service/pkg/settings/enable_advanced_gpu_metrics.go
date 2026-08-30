package settings

import (
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// EnableAdvancedGPUMetrics controls whether advanced GPU metrics dashboards/queries are enabled for the tenant.
type EnableAdvancedGPUMetrics struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableAdvancedGPUMetrics) Category() string {
	return ClusterCategory
}

func (e *EnableAdvancedGPUMetrics) Description() string {
	return "Enable advanced GPU metrics collection and dashboards."
}

func (e *EnableAdvancedGPUMetrics) DisplayName() string {
	return "Advanced GPU metrics"
}

func (e *EnableAdvancedGPUMetrics) Key() string {
	return string(api.CLUSTERSETTINGSKEYENUM_ENABLE_ADVANCED_GPU_METRICS)
}

func (e *EnableAdvancedGPUMetrics) SetDefault() {
	e.Enabled = false
}

func (e *EnableAdvancedGPUMetrics) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *EnableAdvancedGPUMetrics) Validate(value interface{}) bool {
	return true
}

func (e *EnableAdvancedGPUMetrics) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return EnableAdvancedGPUMetrics{Enabled: settingValue}
}
