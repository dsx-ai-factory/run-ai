/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiAllocatedMemoryBytesHoursPerDepartmentMetric struct {
	Filters        RunaiAllocatedMemoryBytesHoursPerDepartmentFilters
	ResponseLabels RunaiAllocatedMemoryBytesHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) GetName() string {
	return "runai_allocated_memory_bytes_hours_per_department"
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiAllocatedMemoryBytesHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
