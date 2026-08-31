/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimGpuCacheUsagePercPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimGpuCacheUsagePercPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimGpuCacheUsagePercPerPodMetric struct {
	Filters        RunaiNimGpuCacheUsagePercPerPodFilters
	ResponseLabels RunaiNimGpuCacheUsagePercPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) GetName() string {
	return "runai_nim_gpu_cache_usage_perc_per_pod"
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimGpuCacheUsagePercPerPodMetric) IsQuantile() bool {
	return false
}
