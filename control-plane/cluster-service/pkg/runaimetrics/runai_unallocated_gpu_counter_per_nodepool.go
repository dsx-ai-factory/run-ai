/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiUnallocatedGpuCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiUnallocatedGpuCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiUnallocatedGpuCountPerNodepoolMetric struct {
	Filters        RunaiUnallocatedGpuCountPerNodepoolFilters
	ResponseLabels RunaiUnallocatedGpuCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) GetName() string {
	return "runai_unallocated_gpu_count_per_nodepool"
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiUnallocatedGpuCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
