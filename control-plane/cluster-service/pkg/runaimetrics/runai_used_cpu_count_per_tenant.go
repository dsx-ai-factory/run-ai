/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresPerTenantMetric struct {
	Filters        RunaiUsedCpuCoresPerTenantFilters
	ResponseLabels RunaiUsedCpuCoresPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresPerTenantMetric) GetName() string {
	return "runai_used_cpu_cores_per_tenant"
}

func (m RunaiUsedCpuCoresPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedCpuCoresPerTenantMetric) IsQuantile() bool {
	return false
}
