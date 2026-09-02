/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestsWaitingPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestsWaitingPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestsWaitingPerPodMetric struct {
	Filters        RunaiNimNumRequestsWaitingPerPodFilters
	ResponseLabels RunaiNimNumRequestsWaitingPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) GetName() string {
	return "runai_nim_num_requests_waiting_per_pod"
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestsWaitingPerPodMetric) IsQuantile() bool {
	return false
}
