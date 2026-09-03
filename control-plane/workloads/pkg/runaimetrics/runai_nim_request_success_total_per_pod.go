/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimRequestSuccessTotalPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimRequestSuccessTotalPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimRequestSuccessTotalPerPodMetric struct {
	Filters        RunaiNimRequestSuccessTotalPerPodFilters
	ResponseLabels RunaiNimRequestSuccessTotalPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) GetName() string {
	return "runai_nim_request_success_total_per_pod"
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimRequestSuccessTotalPerPodMetric) IsQuantile() bool {
	return false
}
