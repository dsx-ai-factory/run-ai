/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimRequestFailureTotalPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimRequestFailureTotalPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimRequestFailureTotalPerPodMetric struct {
	Filters        RunaiNimRequestFailureTotalPerPodFilters
	ResponseLabels RunaiNimRequestFailureTotalPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimRequestFailureTotalPerPodMetric) GetName() string {
	return "runai_nim_request_failure_total_per_pod"
}

func (m RunaiNimRequestFailureTotalPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimRequestFailureTotalPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimRequestFailureTotalPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimRequestFailureTotalPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimRequestFailureTotalPerPodMetric) IsQuantile() bool {
	return false
}
