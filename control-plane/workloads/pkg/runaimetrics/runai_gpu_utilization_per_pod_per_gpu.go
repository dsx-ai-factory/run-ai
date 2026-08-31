/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiGpuUtilizationPerPodPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiGpuUtilizationPerPodPerGpuResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
	Gpu   bool `mapstructure:"gpu,omitempty"`
}

type RunaiGpuUtilizationPerPodPerGpuMetric struct {
	Filters        RunaiGpuUtilizationPerPodPerGpuFilters
	ResponseLabels RunaiGpuUtilizationPerPodPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) GetName() string {
	return "runai_gpu_utilization_per_pod_per_gpu"
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiGpuUtilizationPerPodPerGpuMetric) IsQuantile() bool {
	return false
}
