/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiFreeGpuCountPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiFreeGpuCountPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiFreeGpuCountPerNodepoolMetric struct {
	Filters        RunaiFreeGpuCountPerNodepoolFilters
	ResponseLabels RunaiFreeGpuCountPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiFreeGpuCountPerNodepoolMetric) GetName() string {
	return "runai_free_gpu_count_per_nodepool"
}

func (m RunaiFreeGpuCountPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiFreeGpuCountPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiFreeGpuCountPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiFreeGpuCountPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiFreeGpuCountPerNodepoolMetric) IsQuantile() bool {
	return false
}
