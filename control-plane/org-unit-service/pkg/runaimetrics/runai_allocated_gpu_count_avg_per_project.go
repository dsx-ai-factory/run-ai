/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountAvgPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuCountAvgPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuCountAvgPerProjectMetric struct {
	Filters        RunaiAllocatedGpuCountAvgPerProjectFilters
	ResponseLabels RunaiAllocatedGpuCountAvgPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) GetName() string {
	return "runai_allocated_gpu_count_avg_per_project"
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuCountAvgPerProjectMetric) IsQuantile() bool {
	return false
}
