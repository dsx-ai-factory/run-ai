/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiNimNumRequestMaxPerPodFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId     []string    `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestMaxPerPodResponseLabels struct {
	PodId bool `mapstructure:"pod_uuid,omitempty"`
}

type RunaiNimNumRequestMaxPerPodMetric struct {
	Filters        RunaiNimNumRequestMaxPerPodFilters
	ResponseLabels RunaiNimNumRequestMaxPerPodResponseLabels
	Quantile       *float32
}

func (m RunaiNimNumRequestMaxPerPodMetric) GetName() string {
	return "runai_nim_num_request_max_per_pod"
}

func (m RunaiNimNumRequestMaxPerPodMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiNimNumRequestMaxPerPodMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiNimNumRequestMaxPerPodMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiNimNumRequestMaxPerPodMetric) IsAggregation() bool {
	return false
}

func (m RunaiNimNumRequestMaxPerPodMetric) IsQuantile() bool {
	return false
}
