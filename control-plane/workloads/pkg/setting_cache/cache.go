package setting_cache

import (
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type AdvancedGpuMetricsSettingCacheValue struct {
	IsOn bool
}

func RegisterAdvancedGpuMetricsStore() {
	err := cache.RegisterStore[AdvancedGpuMetricsSettingCacheValue](options.Ttl(30 * time.Minute))
	if err != nil {
		logging.Warnf("Failed to register advanced GPU metrics settings cache store: %v", err)
	}
}
