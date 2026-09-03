/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiDeservedCpuMillicoresPerQueuesFilters struct {
	ClusterId      []string `mapstructure:"clusterId,omitempty"`
	Project        []string `mapstructure:"project,omitempty"`
	NodePool       []string `mapstructure:"nodepool,omitempty"`
	Department     []string `mapstructure:"department,omitempty"`
	DepartmentName []string `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedCpuMillicoresPerQueuesResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	Project        bool `mapstructure:"project,omitempty"`
	Nodepool       bool `mapstructure:"nodepool,omitempty"`
	Department     bool `mapstructure:"department,omitempty"`
	DepartmentName bool `mapstructure:"department_name,omitempty"`
}

type RunaiDeservedCpuMillicoresPerQueuesMetric struct {
	Filters        RunaiDeservedCpuMillicoresPerQueuesFilters
	ResponseLabels RunaiDeservedCpuMillicoresPerQueuesResponseLabels
	Quantile       *float32
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) GetName() string {
	return "runai_deserved_cpu_millicores_per_queues"
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) IsAggregation() bool {
	return false
}

func (m RunaiDeservedCpuMillicoresPerQueuesMetric) IsQuantile() bool {
	return false
}
