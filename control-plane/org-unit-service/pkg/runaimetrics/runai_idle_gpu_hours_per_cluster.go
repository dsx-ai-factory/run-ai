/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleGpuHoursPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiIdleGpuHoursPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiIdleGpuHoursPerClusterMetric struct {
	Filters        RunaiIdleGpuHoursPerClusterFilters
	ResponseLabels RunaiIdleGpuHoursPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiIdleGpuHoursPerClusterMetric) GetName() string {
	return "runai_idle_gpu_hours_per_cluster"
}

func (m RunaiIdleGpuHoursPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleGpuHoursPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleGpuHoursPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleGpuHoursPerClusterMetric) IsAggregation() bool {
	return true
}

func (m RunaiIdleGpuHoursPerClusterMetric) IsQuantile() bool {
	return false
}
