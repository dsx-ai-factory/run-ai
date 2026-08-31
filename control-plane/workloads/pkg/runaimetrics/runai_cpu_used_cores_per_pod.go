/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiCpuUsedCoresPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiCpuUsedCoresPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiCpuUsedCoresPerPodMetric struct {
	Filters        RunaiCpuUsedCoresPerPodFilters
	ResponseLabels RunaiCpuUsedCoresPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiCpuUsedCoresPerPodMetric) GetName() string {
	return "runai_cpu_used_cores_per_pod"
}

func (m RunaiCpuUsedCoresPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiCpuUsedCoresPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiCpuUsedCoresPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiCpuUsedCoresPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiCpuUsedCoresPerPodMetric) IsQuantile() bool {
	return false
}
