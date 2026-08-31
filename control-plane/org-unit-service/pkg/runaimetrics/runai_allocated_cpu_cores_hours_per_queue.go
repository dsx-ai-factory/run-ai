/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiAllocatedCpuCoresHoursPerQueueMetric struct {
	Filters        RunaiAllocatedCpuCoresHoursPerQueueFilters
	ResponseLabels RunaiAllocatedCpuCoresHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) GetName() string {
	return "runai_allocated_cpu_cores_hours_per_queue"
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedCpuCoresHoursPerQueueMetric) IsQuantile() bool {
	return false
}
