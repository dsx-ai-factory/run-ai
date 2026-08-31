/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleAllocatedGpuCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiIdleAllocatedGpuCountPerTenantResponseLabels struct{}

type RunaiIdleAllocatedGpuCountPerTenantMetric struct {
	Filters        RunaiIdleAllocatedGpuCountPerTenantFilters
	ResponseLabels RunaiIdleAllocatedGpuCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) GetName() string {
	return "runai_idle_allocated_gpu_count_per_tenant"
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleAllocatedGpuCountPerTenantMetric) IsQuantile() bool {
	return false
}
