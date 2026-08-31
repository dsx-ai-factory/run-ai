/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesHoursPerProjectFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerProjectResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerProjectMetric struct {
	Filters        RunaiUsedMemoryBytesHoursPerProjectFilters
	ResponseLabels RunaiUsedMemoryBytesHoursPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) GetName() string {
	return "runai_used_memory_bytes_hours_per_project"
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedMemoryBytesHoursPerProjectMetric) IsQuantile() bool {
	return false
}
