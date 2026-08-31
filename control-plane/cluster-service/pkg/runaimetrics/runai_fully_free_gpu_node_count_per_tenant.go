/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFullyFreeGpuNodeCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerTenantResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerTenantMetric struct {
	Filters        RunaiFullyFreeGpuNodeCountPerTenantFilters
	ResponseLabels RunaiFullyFreeGpuNodeCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) GetName() string {
	return "runai_fully_free_gpu_node_count_per_tenant"
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiFullyFreeGpuNodeCountPerTenantMetric) IsQuantile() bool {
	return false
}
