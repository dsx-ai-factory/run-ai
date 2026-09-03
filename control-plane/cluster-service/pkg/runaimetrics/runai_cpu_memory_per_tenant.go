/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiTotalMemoryBytesPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiTotalMemoryBytesPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiTotalMemoryBytesPerTenantMetric struct {
	Filters        RunaiTotalMemoryBytesPerTenantFilters
	ResponseLabels RunaiTotalMemoryBytesPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiTotalMemoryBytesPerTenantMetric) GetName() string {
	return "runai_total_memory_bytes_per_tenant"
}

func (m RunaiTotalMemoryBytesPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiTotalMemoryBytesPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiTotalMemoryBytesPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiTotalMemoryBytesPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiTotalMemoryBytesPerTenantMetric) IsQuantile() bool {
	return false
}
