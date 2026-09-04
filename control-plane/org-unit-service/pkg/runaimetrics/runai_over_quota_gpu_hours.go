/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/google/uuid"
)

type RunaiOverQuotaGpuHoursFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiOverQuotaGpuHoursResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiOverQuotaGpuHoursMetric struct {
	Filters        RunaiOverQuotaGpuHoursFilters
	ResponseLabels RunaiOverQuotaGpuHoursResponseLabels
	Quantile       *float32
	GroupBy        client.MetricGroupBy
}

func (m RunaiOverQuotaGpuHoursMetric) GetName() string {
	switch m.GroupBy {
	case client.GroupByCluster:
		return "runai_over_quota_gpu_hours_per_cluster"
	case client.GroupByNodePool:
		return "runai_over_quota_gpu_hours_per_nodepool"
	case client.GroupByProject:
		return "runai_over_quota_gpu_hours_per_project"
	case client.GroupByDepartment:
		return "runai_over_quota_gpu_hours_per_department"
	}

	return "runai_over_quota_gpu_hours_per_cluster"
}

func (m RunaiOverQuotaGpuHoursMetric) GetFilters() any {
	return m.Filters
}

func (m RunaiOverQuotaGpuHoursMetric) GetResponseLabels() any {
	return m.ResponseLabels
}

func (m RunaiOverQuotaGpuHoursMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiOverQuotaGpuHoursMetric) IsAggregation() bool {
	return true
}

func (m RunaiOverQuotaGpuHoursMetric) IsQuantile() bool {
	return false
}

func (m RunaiOverQuotaGpuHoursMetric) GetGroupBy() client.MetricGroupBy {
	return m.GroupBy
}
