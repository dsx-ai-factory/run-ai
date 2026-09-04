/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedMemoryBytesHoursPerQueueMetric struct {
	Filters        RunaiUsedMemoryBytesHoursPerQueueFilters
	ResponseLabels RunaiUsedMemoryBytesHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) GetName() string {
	return "runai_used_memory_bytes_hours_per_queue"
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedMemoryBytesHoursPerQueueMetric) IsQuantile() bool {
	return false
}
