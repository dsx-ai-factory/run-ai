/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

type RunaiInQuotaGpuHoursFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiInQuotaGpuHoursResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiInQuotaGpuHoursMetric struct {
	Filters        RunaiInQuotaGpuHoursFilters
	ResponseLabels RunaiInQuotaGpuHoursResponseLabels
	Quantile       *float32
	GroupBy        client.MetricGroupBy
}

func (m RunaiInQuotaGpuHoursMetric) GetName() string {
	switch m.GroupBy {
	case client.GroupByCluster:
		return "runai_in_quota_gpu_hours_per_cluster"
	case client.GroupByNodePool:
		return "runai_in_quota_gpu_hours_per_nodepool"
	case client.GroupByProject:
		return "runai_in_quota_gpu_hours_per_project"
	case client.GroupByDepartment:
		return "runai_in_quota_gpu_hours_per_department"
	}

	return "runai_in_quota_gpu_hours_per_cluster"
}

func (m RunaiInQuotaGpuHoursMetric) GetFilters() any {
	return m.Filters
}

func (m RunaiInQuotaGpuHoursMetric) GetResponseLabels() any {
	return m.ResponseLabels
}

func (m RunaiInQuotaGpuHoursMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiInQuotaGpuHoursMetric) IsAggregation() bool {
	return true
}

func (m RunaiInQuotaGpuHoursMetric) IsQuantile() bool {
	return false
}

func (m RunaiInQuotaGpuHoursMetric) GetGroupBy() client.MetricGroupBy {
	return m.GroupBy
}
