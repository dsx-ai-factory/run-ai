/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiMemoryUtilizationPerNodepoolsFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
	NodePool  []string `mapstructure:"nodepool,omitempty"`
}

type RunaiMemoryUtilizationPerNodepoolsResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
}

type RunaiMemoryUtilizationPerNodepoolsMetric struct {
	Filters        RunaiMemoryUtilizationPerNodepoolsFilters
	ResponseLabels RunaiMemoryUtilizationPerNodepoolsResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) GetName() string {
	return "runai_memory_utilization_per_nodepools"
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUtilizationPerNodepoolsMetric) IsQuantile() bool {
	return false
}
