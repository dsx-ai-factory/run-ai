/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

type RunaiWorkloadsCountPerQueuesFilters struct {
	ClusterId  []string `mapstructure:"clusterId,omitempty"`
	Project    []string `mapstructure:"project,omitempty"`
	NodePool   []string `mapstructure:"nodepool,omitempty"`
	Department []string `mapstructure:"department,omitempty"`
}

type RunaiWorkloadsCountPerQueuesResponseLabels struct {
	ClusterId  bool `mapstructure:"clusterId,omitempty"`
	Project    bool `mapstructure:"project,omitempty"`
	Nodepool   bool `mapstructure:"nodepool,omitempty"`
	Department bool `mapstructure:"department,omitempty"`
}

type RunaiWorkloadsCountPerQueuesMetric struct {
	Filters        RunaiWorkloadsCountPerQueuesFilters
	ResponseLabels RunaiWorkloadsCountPerQueuesResponseLabels
	Quantile       *float32
	GroupBy        client.MetricGroupBy
}

func (m RunaiWorkloadsCountPerQueuesMetric) GetName() string {
	switch m.GroupBy {
	case client.GroupByProject:
		return "runai_workload_count_per_project"
	case client.GroupByDepartment:
		return "runai_workload_count_per_department"
	case client.GroupByCategory:
		return "runai_workload_count_per_category"
	}

	return "runai_workload_count_per_department"
}

func (m RunaiWorkloadsCountPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiWorkloadsCountPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiWorkloadsCountPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiWorkloadsCountPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiWorkloadsCountPerQueuesMetric) IsQuantile() bool {
	return false
}

func (m RunaiWorkloadsCountPerQueuesMetric) GetGroupBy() client.MetricGroupBy {
	return m.GroupBy
}
