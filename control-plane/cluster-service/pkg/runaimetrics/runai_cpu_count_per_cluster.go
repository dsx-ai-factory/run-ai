/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuCoresPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuCoresPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiCpuCoresPerClusterMetric struct {
	Filters        RunaiCpuCoresPerClusterFilters
	ResponseLabels RunaiCpuCoresPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiCpuCoresPerClusterMetric) GetName() string {
	return "runai_total_cpu_cores_per_cluster"
}

func (m RunaiCpuCoresPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuCoresPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuCoresPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuCoresPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuCoresPerClusterMetric) IsQuantile() bool {
	return false
}
