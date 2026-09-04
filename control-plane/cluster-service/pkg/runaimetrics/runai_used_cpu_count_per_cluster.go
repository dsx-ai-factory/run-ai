/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUsedCpuCoresPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiUsedCpuCoresPerClusterMetric struct {
	Filters        RunaiUsedCpuCoresPerClusterFilters
	ResponseLabels RunaiUsedCpuCoresPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiUsedCpuCoresPerClusterMetric) GetName() string {
	return "runai_used_cpu_cores_per_cluster"
}

func (m RunaiUsedCpuCoresPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUsedCpuCoresPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUsedCpuCoresPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUsedCpuCoresPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiUsedCpuCoresPerClusterMetric) IsQuantile() bool {
	return false
}
