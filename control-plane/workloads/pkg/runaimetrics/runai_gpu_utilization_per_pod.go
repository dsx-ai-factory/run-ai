/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuUtilizationPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuUtilizationPerPodMetric struct {
	Filters        RunaiGpuUtilizationPerPodFilters
	ResponseLabels RunaiGpuUtilizationPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerPodMetric) GetName() string {
	return "runai_gpu_utilization_per_pod"
}

func (m RunaiGpuUtilizationPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerPodMetric) IsQuantile() bool {
	return false
}
