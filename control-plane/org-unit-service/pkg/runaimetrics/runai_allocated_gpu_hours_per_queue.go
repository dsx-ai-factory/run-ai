/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedGpuHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedGpuHoursPerQueueMetric struct {
	Filters        RunaiAllocatedGpuHoursPerQueueFilters
	ResponseLabels RunaiAllocatedGpuHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) GetName() string {
	return "runai_allocated_gpu_hours_per_queue"
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedGpuHoursPerQueueMetric) IsQuantile() bool {
	return false
}
