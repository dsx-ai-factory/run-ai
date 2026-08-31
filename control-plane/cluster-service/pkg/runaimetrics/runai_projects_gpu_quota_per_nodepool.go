/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiProjectsGpuQuotaPerNodepoolFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiProjectsGpuQuotaPerNodepoolResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiProjectsGpuQuotaPerNodepoolMetric struct {
	Filters        RunaiProjectsGpuQuotaPerNodepoolFilters
	ResponseLabels RunaiProjectsGpuQuotaPerNodepoolResponseLabels
	Quantile       *float32
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) GetName() string {
	return "runai_projects_gpu_quota_per_nodepool"
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) IsAggregation() bool {
	return false
}

func (m RunaiProjectsGpuQuotaPerNodepoolMetric) IsQuantile() bool {
	return false
}
