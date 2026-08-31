/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuMemoryBytesPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryBytesPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuMemoryBytesPerTenantMetric struct {
	Filters        RunaiGpuMemoryBytesPerTenantFilters
	ResponseLabels RunaiGpuMemoryBytesPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiGpuMemoryBytesPerTenantMetric) GetName() string {
	return "runai_total_gpu_memory_byte_per_tenant"
}

func (m RunaiGpuMemoryBytesPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuMemoryBytesPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuMemoryBytesPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuMemoryBytesPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuMemoryBytesPerTenantMetric) IsQuantile() bool {
	return false
}
