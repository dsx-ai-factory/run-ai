/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuUtilizationPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuUtilizationPerNodepoolMetric struct {
	Filters        RunaiGpuUtilizationPerNodepoolFilters
	ResponseLabels RunaiGpuUtilizationPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerNodepoolMetric) GetName() string {
	return "runai_gpu_utilization_per_nodepool"
}

func (m RunaiGpuUtilizationPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerNodepoolMetric) IsQuantile() bool {
	return false
}
