/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedMemoryBytesPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedMemoryBytesPerTenantMetric struct {
	Filters        RunaiAllocatedMemoryBytesPerTenantFilters
	ResponseLabels RunaiAllocatedMemoryBytesPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) GetName() string {
	return "runai_allocated_memory_bytes_per_tenant"
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedMemoryBytesPerTenantMetric) IsQuantile() bool {
	return false
}
