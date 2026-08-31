/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuCoresPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuCoresPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuCoresPerNodepoolMetric struct {
	Filters        RunaiCpuCoresPerNodepoolFilters
	ResponseLabels RunaiCpuCoresPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiCpuCoresPerNodepoolMetric) GetName() string {
	return "runai_total_cpu_cores_per_nodepool"
}

func (m RunaiCpuCoresPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuCoresPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuCoresPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuCoresPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuCoresPerNodepoolMetric) IsQuantile() bool {
	return false
}
