/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedMemoryBytesPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedMemoryBytesPerTenantMetric struct {
	Filters        RunaiUsedMemoryBytesPerTenantFilters
	ResponseLabels RunaiUsedMemoryBytesPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiUsedMemoryBytesPerTenantMetric) GetName() string {
	return "runai_used_memory_bytes_per_tenant"
}

func (m RunaiUsedMemoryBytesPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedMemoryBytesPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedMemoryBytesPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedMemoryBytesPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedMemoryBytesPerTenantMetric) IsQuantile() bool {
	return false
}
