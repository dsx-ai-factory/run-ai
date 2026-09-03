/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuUtilizationPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiGpuUtilizationPerClusterMetric struct {
	Filters        RunaiGpuUtilizationPerClusterFilters
	ResponseLabels RunaiGpuUtilizationPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerClusterMetric) GetName() string {
	return "runai_gpu_utilization_per_cluster"
}

func (m RunaiGpuUtilizationPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerClusterMetric) IsQuantile() bool {
	return false
}
