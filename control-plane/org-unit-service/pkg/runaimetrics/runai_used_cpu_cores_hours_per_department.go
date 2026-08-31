/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresHoursPerDepartmentFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
}

type RunaiUsedCpuCoresHoursPerDepartmentResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiUsedCpuCoresHoursPerDepartmentMetric struct {
	Filters        RunaiUsedCpuCoresHoursPerDepartmentFilters
	ResponseLabels RunaiUsedCpuCoresHoursPerDepartmentResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) GetName() string {
	return "runai_used_cpu_cores_hours_per_department"
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) IsAggregation() bool {
	return true
}

func (m RunaiUsedCpuCoresHoursPerDepartmentMetric) IsQuantile() bool {
	return false
}
