/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

type RunaiWorkloadsCountFilters struct {
	ClusterId  []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Department []string    `mapstructure:"department,omitempty"`
	Project    []string    `mapstructure:"project,omitempty"`
	Nodepool   []string    `mapstructure:"nodepool,omitempty"`
}

type RunaiWorkloadsCountResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
}

type RunaiWorkloadsCountMetric struct {
	Filters        RunaiWorkloadsCountFilters
	ResponseLabels RunaiWorkloadsCountResponseLabels
	Quantile       *float32
	GroupBy        client.MetricGroupBy
}

func (m RunaiWorkloadsCountMetric) GetName() string {
	switch m.GroupBy {
	case client.GroupByCluster:
		return "runai_workload_count_per_cluster"
	case client.GroupByCategory:
		return "runai_workload_count_per_category"
	case client.GroupByNodePool:
		return "runai_workload_count_per_nodepool"
	case client.GroupByProject:
		return "runai_workload_count_per_project"
	case client.GroupByDepartment:
		return "runai_workload_count_per_department"
	}

	return "runai_workload_count_per_tenant"
}

func (m RunaiWorkloadsCountMetric) GetFilters() any {
	return m.Filters
}

func (m RunaiWorkloadsCountMetric) GetResponseLabels() any {
	return m.ResponseLabels
}

func (m RunaiWorkloadsCountMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiWorkloadsCountMetric) IsAggregation() bool {
	return false
}

func (m RunaiWorkloadsCountMetric) IsQuantile() bool {
	return false
}

func (m RunaiWorkloadsCountMetric) GetGroupBy() client.MetricGroupBy {
	return m.GroupBy
}
