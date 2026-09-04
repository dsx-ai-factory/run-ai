/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFullyFreeGpuNodeCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiFullyFreeGpuNodeCountPerClusterMetric struct {
	Filters        RunaiFullyFreeGpuNodeCountPerClusterFilters
	ResponseLabels RunaiFullyFreeGpuNodeCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) GetName() string {
	return "runai_fully_free_gpu_node_count_per_cluster"
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiFullyFreeGpuNodeCountPerClusterMetric) IsQuantile() bool {
	return false
}
