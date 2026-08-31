/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerDepartmentMetric struct {
	Filters        RunaiUsedMemoryBytesHoursPerDepartmentFilters
	ResponseLabels RunaiUsedMemoryBytesHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) GetName() string {
	return "runai_used_memory_bytes_hours_per_department"
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedMemoryBytesHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
