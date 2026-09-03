/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

type RunaiAllocatedGpuCountPerProjectFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
	Project   []string `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuCountPerProjectResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Project   bool `mapstructure:"project,omitempty"`
}

type RunaiAllocatedGpuCountPerProjectMetric struct {
	Filters        RunaiAllocatedGpuCountPerProjectFilters
	ResponseLabels RunaiAllocatedGpuCountPerProjectResponseLabels
	Quantile       *float32
}

func (m RunaiAllocatedGpuCountPerProjectMetric) GetName() string {
	return "runai_allocated_gpu_count_per_project"
}

func (m RunaiAllocatedGpuCountPerProjectMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiAllocatedGpuCountPerProjectMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiAllocatedGpuCountPerProjectMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiAllocatedGpuCountPerProjectMetric) IsAggregation() bool {
	return false
}

func (m RunaiAllocatedGpuCountPerProjectMetric) IsQuantile() bool {
	return false
}
