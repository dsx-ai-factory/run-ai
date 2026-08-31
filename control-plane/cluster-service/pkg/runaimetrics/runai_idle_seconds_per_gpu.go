/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiIdleSecondsPerGpuFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Nodepool  []string    `mapstructure:"nodepool,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
}

type RunaiIdleSecondsPerGpuResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Nodepool  bool `mapstructure:"nodepool,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiIdleSecondsPerGpuMetric struct {
	Filters        RunaiIdleSecondsPerGpuFilters
	ResponseLabels RunaiIdleSecondsPerGpuResponseLabels
	Quantile       *float32
}

func (m RunaiIdleSecondsPerGpuMetric) GetName() string {
	return "runai_idle_seconds_per_gpu"
}

func (m RunaiIdleSecondsPerGpuMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiIdleSecondsPerGpuMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiIdleSecondsPerGpuMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiIdleSecondsPerGpuMetric) IsAggregation() bool {
	return false
}

func (m RunaiIdleSecondsPerGpuMetric) IsQuantile() bool {
	return false
}
