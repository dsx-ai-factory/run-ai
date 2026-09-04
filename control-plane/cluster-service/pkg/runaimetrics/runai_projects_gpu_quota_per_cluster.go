/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiProjectsGpuQuotaPerClusterFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
}

type RunaiProjectsGpuQuotaPerClusterResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type RunaiProjectsGpuQuotaPerClusterMetric struct {
	Filters        RunaiProjectsGpuQuotaPerClusterFilters
	ResponseLabels RunaiProjectsGpuQuotaPerClusterResponseLabels
	Quantile       *float32
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) GetName() string {
	return "runai_projects_gpu_quota_per_cluster"
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) IsAggregation() bool {
	return false
}

func (m RunaiProjectsGpuQuotaPerClusterMetric) IsQuantile() bool {
	return false
}
