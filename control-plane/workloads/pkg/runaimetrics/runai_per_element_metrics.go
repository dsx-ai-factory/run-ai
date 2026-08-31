/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type ElementMetricName string

const (
	ElementGpuUtilization          ElementMetricName = "runai_gpu_utilization_per_element"
	ElementGpuAllocation           ElementMetricName = "runai_gpu_allocation_per_element"
	ElementGpuMemoryUsedBytes      ElementMetricName = "runai_gpu_memory_used_bytes_per_element"
	ElementGpuMemoryRequestedBytes ElementMetricName = "runai_gpu_memory_requested_bytes_per_element"
	ElementCpuUsedCores            ElementMetricName = "runai_cpu_used_cores_per_element"
	ElementCpuRequestedCores       ElementMetricName = "runai_cpu_requested_cores_per_element"
	ElementCpuLimitCores           ElementMetricName = "runai_cpu_limit_cores_per_element"
	ElementMemoryUsedBytes         ElementMetricName = "runai_memory_used_bytes_per_element"
	ElementMemoryRequestedBytes    ElementMetricName = "runai_memory_requested_bytes_per_element"
	ElementMemoryLimitBytes        ElementMetricName = "runai_memory_limit_bytes_per_element"
)

type PerElementFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	ElementId []string    `mapstructure:"element_id,omitempty"`
}

type PerElementMetric struct {
	Name    ElementMetricName
	Filters PerElementFilters
}

func NewPerElementMetric(name ElementMetricName, clusterFilter []uuid.UUID, elementFilter []string) PerElementMetric {
	return PerElementMetric{
		Name:    name,
		Filters: PerElementFilters{ClusterId: clusterFilter, ElementId: elementFilter},
	}
}

func (m PerElementMetric) GetName() string                { return string(m.Name) }
func (m PerElementMetric) GetFilters() interface{}        { return m.Filters }
func (m PerElementMetric) GetResponseLabels() interface{} { return nil }
func (m PerElementMetric) GetQuantile() *float32          { return nil }
func (m PerElementMetric) IsAggregation() bool            { return false }
func (m PerElementMetric) IsQuantile() bool               { return false }
