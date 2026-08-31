/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiAllocatedCpuCoresPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiAllocatedCpuCoresPerClusterMetric struct {
	Filters        RunaiAllocatedCpuCoresPerClusterFilters
	ResponseLabels RunaiAllocatedCpuCoresPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) GetName() string {
	return "runai_allocated_cpu_cores_per_cluster"
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedCpuCoresPerClusterMetric) IsQuantile() bool {
	return false
}
