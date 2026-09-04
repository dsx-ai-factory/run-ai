/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresPerTenantMetric struct {
	Filters        RunaiAllocatedCpuCoresPerTenantFilters
	ResponseLabels RunaiAllocatedCpuCoresPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) GetName() string {
	return "runai_allocated_cpu_cores_per_tenant"
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuCoresPerTenantMetric) IsQuantile() bool {
	return false
}
