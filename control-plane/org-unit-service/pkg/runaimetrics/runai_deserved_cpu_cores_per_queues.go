/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiDeservedCpuPerQueuesFilters struct {
	ClusterId      []string `mapstructure:"clusterId,omitempty"`
	Project        []string `mapstructure:"project,omitempty"`
	NodePool       []string `mapstructure:"nodepool,omitempty"`
	Department     []string `mapstructure:"department,omitempty"`
	DepartmentName []string `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedCpuPerQueuesResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	Project        bool `mapstructure:"project,omitempty"`
	Nodepool       bool `mapstructure:"nodepool,omitempty"`
	Department     bool `mapstructure:"department,omitempty"`
	DepartmentName bool `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedCpuMetric struct {
	Filters        RunaiDeservedCpuPerQueuesFilters
	ResponseLabels RunaiDeservedCpuPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiDeservedCpuMetric) GetName() string {
	return "runai_deserved_cpu_per_queues"
}

func (m RunaiDeservedCpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDeservedCpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDeservedCpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDeservedCpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiDeservedCpuMetric) IsQuantile() bool {
	return false
}
