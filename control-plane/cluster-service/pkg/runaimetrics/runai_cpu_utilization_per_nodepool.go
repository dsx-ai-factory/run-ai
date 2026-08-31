/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuUtilizationPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuUtilizationPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiCpuUtilizationPerNodepoolMetric struct {
	Filters        RunaiCpuUtilizationPerNodepoolFilters
	ResponseLabels RunaiCpuUtilizationPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUtilizationPerNodepoolMetric) GetName() string {
	return "runai_cpu_utilization_per_nodepool"
}

func (m RunaiCpuUtilizationPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUtilizationPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUtilizationPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUtilizationPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUtilizationPerNodepoolMetric) IsQuantile() bool {
	return false
}
