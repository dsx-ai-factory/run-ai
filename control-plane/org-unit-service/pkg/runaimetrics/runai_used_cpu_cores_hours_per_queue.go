/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresHoursPerQueueFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedCpuCoresHoursPerQueueResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUsedCpuCoresHoursPerQueueMetric struct {
	Filters        RunaiUsedCpuCoresHoursPerQueueFilters
	ResponseLabels RunaiUsedCpuCoresHoursPerQueueResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) GetName() string {
	return "runai_used_cpu_cores_hours_per_queue"
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedCpuCoresHoursPerQueueMetric) IsQuantile() bool {
	return false
}
