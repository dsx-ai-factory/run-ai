/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuHoursPerProjectMetric struct {
	Filters        RunaiAllocatedGpuHoursPerProjectFilters
	ResponseLabels RunaiAllocatedGpuHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) GetName() string {
	return "runai_allocated_gpu_hours_per_project"
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuHoursPerProjectMetric) IsQuantile() bool {
	return false
}
