/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuNodesCountPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuNodesCountPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuNodesCountPerClusterMetric struct {
	Filters        RunaiGpuNodesCountPerClusterFilters
	ResponseLabels RunaiGpuNodesCountPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiGpuNodesCountPerClusterMetric) GetName() string {
	return "runai_gpu_nodes_count_per_cluster"
}

func (m RunaiGpuNodesCountPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuNodesCountPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuNodesCountPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuNodesCountPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuNodesCountPerClusterMetric) IsQuantile() bool {
	return false
}
