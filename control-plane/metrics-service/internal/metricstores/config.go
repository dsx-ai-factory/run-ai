/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

type MetricStoresConfiguration struct {
	Cache *CacheConfiguration `mapstructure:"cache"`
}

type CacheConfiguration struct {
	TTLMinutes int `mapstructure:"ttl_minutes" validate:"required,min=1"`
}

func (mc *MetricStoresConfiguration) SetDefaults() {
	mc.Cache = &CacheConfiguration{
		TTLMinutes: 60 * 24, // 1 day
	}
}

type DatasourceIdConfig struct {
	Env     string `mapstructure:"env"`
	Release string `mapstructure:"release"`
}

func (c *DatasourceIdConfig) SetDefaults() {}
