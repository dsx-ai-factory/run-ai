/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationDistributionPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiGpuUtilizationDistributionPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Range     bool `mapstructure:"range,omitempty"`
}

type RunaiGpuUtilizationDistributionPerNodepoolMetric struct {
	Filters        RunaiGpuUtilizationDistributionPerNodepoolFilters
	ResponseLabels RunaiGpuUtilizationDistributionPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) GetName() string {
	return "runai_gpu_utilization_distribution_per_nodepool"
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationDistributionPerNodepoolMetric) IsQuantile() bool {
	return false
}
