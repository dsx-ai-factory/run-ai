/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setting_cache

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
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
