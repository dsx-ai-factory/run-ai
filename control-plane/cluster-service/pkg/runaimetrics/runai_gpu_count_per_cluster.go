/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuCountPerClusterMetric struct {
	Filters        RunaiGpuCountPerClusterFilters
	ResponseLabels RunaiGpuCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiGpuCountPerClusterMetric) GetName() string {
	return "runai_gpu_count_per_cluster"
}

func (m RunaiGpuCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuCountPerClusterMetric) IsQuantile() bool {
	return false
}
