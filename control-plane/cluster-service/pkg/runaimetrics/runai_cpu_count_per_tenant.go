/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuCoresPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuCoresPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuCoresPerTenantMetric struct {
	Filters        RunaiCpuCoresPerTenantFilters
	ResponseLabels RunaiCpuCoresPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiCpuCoresPerTenantMetric) GetName() string {
	return "runai_total_cpu_cores_per_tenant"
}

func (m RunaiCpuCoresPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuCoresPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuCoresPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuCoresPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuCoresPerTenantMetric) IsQuantile() bool {
	return false
}
