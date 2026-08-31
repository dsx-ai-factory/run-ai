/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedGpuCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedGpuCountPerTenantResponseLabels struct{}

type RunaiAllocatedGpuCountPerTenantMetric struct {
	Filters        RunaiAllocatedGpuCountPerTenantFilters
	ResponseLabels RunaiAllocatedGpuCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerTenantMetric) GetName() string {
	return "runai_allocated_gpu_count_per_tenant"
}

func (m RunaiAllocatedGpuCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerTenantMetric) IsQuantile() bool {
	return false
}
