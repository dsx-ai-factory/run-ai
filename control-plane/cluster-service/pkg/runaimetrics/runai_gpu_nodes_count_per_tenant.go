/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuNodesCountPerTenantFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuNodesCountPerTenantResponseLabels struct{}

type RunaiGpuNodesCountPerTenantMetric struct {
	Filters        RunaiGpuNodesCountPerTenantFilters
	ResponseLabels RunaiGpuNodesCountPerTenantResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNodesCountPerTenantMetric) GetName() string {
	return "runai_gpu_nodes_count_per_tenant"
}

func (m RunaiGpuNodesCountPerTenantMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNodesCountPerTenantMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNodesCountPerTenantMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNodesCountPerTenantMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNodesCountPerTenantMetric) IsQuantile() bool {
	return false
}
