/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiDeservedMemoryMbPerQueuesFilters struct {
	ClusterId      []string `mapstructure:"clusterId,omitempty"`
	Project        []string `mapstructure:"project,omitempty"`
	NodePool       []string `mapstructure:"nodepool,omitempty"`
	Department     []string `mapstructure:"department,omitempty"`
	DepartmentName []string `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedMemoryMbPerQueuesResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	Project        bool `mapstructure:"project,omitempty"`
	Nodepool       bool `mapstructure:"nodepool,omitempty"`
	Department     bool `mapstructure:"department,omitempty"`
	DepartmentName bool `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedMemoryMbPerQueuesMetric struct {
	Filters        RunaiDeservedMemoryMbPerQueuesFilters
	ResponseLabels RunaiDeservedMemoryMbPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) GetName() string {
	return "runai_deserved_memory_mb_per_queues"
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiDeservedMemoryMbPerQueuesMetric) IsQuantile() bool {
	return false
}
