/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"

type MetricGroupBy string

const (
	GroupByTenant     MetricGroupBy = "groupByTenant"
	GroupByCluster    MetricGroupBy = "groupByCluster"
	GroupByNodePool   MetricGroupBy = "groupByNodePool"
	GroupByNode       MetricGroupBy = "groupByNode"
	GroupByCategory   MetricGroupBy = "groupByCategory"
	GroupByProject    MetricGroupBy = "groupByProject"
	GroupByDepartment MetricGroupBy = "groupByDepartment"

	ClusterId = "ClusterId"
	Nodepool  = "Nodepool"
	Node      = "Node"
	Category  = "Category"
)

func (m MetricGroupBy) ToClientGroupBy() client.MetricGroupBy {
	switch m {
	case GroupByCluster:
		return client.GroupByCluster
	case GroupByCategory:
		return client.GroupByCategory
	case GroupByNodePool:
		return client.GroupByNodePool
	case GroupByProject:
		return client.GroupByProject
	case GroupByDepartment:
		return client.GroupByDepartment
	case GroupByNode:
		return client.GroupByNode
	}
	return ""
}
