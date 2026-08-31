/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiMemoryUtilizationPerNodeFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiMemoryUtilizationPerNodeResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
}

type RunaiMemoryUtilizationPerNodeMetric struct {
	Filters        RunaiMemoryUtilizationPerNodeFilters
	ResponseLabels RunaiMemoryUtilizationPerNodeResponseLabels
	Quantile       *float32
}

func (m RunaiMemoryUtilizationPerNodeMetric) GetName() string {
	return "runai_memory_utilization_per_node"
}

func (m RunaiMemoryUtilizationPerNodeMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiMemoryUtilizationPerNodeMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiMemoryUtilizationPerNodeMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiMemoryUtilizationPerNodeMetric) IsAggregation() bool {
	return false
}

func (m RunaiMemoryUtilizationPerNodeMetric) IsQuantile() bool {
	return false
}
