/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
)

type RunaiExternalPodInfoFilters struct {
	ClusterId      []uuid.UUID `mapstructure:"clusterId,omitempty"`
	PodId          []string    `mapstructure:"pod_uuid,omitempty"`
	WorkloadId     []string    `mapstructure:"workload_id,omitempty"`
	DetailedStatus []string    `mapstructure:"detailed_status,omitempty"`
	Phase          []string    `mapstructure:"phase,omitempty"`
	Node           []string    `mapstructure:"node,omitempty"`
	GPU            []string    `mapstructure:"gpu,omitempty"`
}

type RunaiExternalPodInfoResponseLabels struct {
	ClusterId      bool `mapstructure:"clusterId,omitempty"`
	PodId          bool `mapstructure:"pod_uuid,omitempty"`
	WorkloadId     bool `mapstructure:"workload_id,omitempty"`
	DetailedStatus bool `mapstructure:"detailed_status,omitempty"`
	Phase          bool `mapstructure:"phase,omitempty"`
	Node           bool `mapstructure:"node,omitempty"`
	GPU            bool `mapstructure:"gpu,omitempty"`
}

type RunaiExternalPodInfoMetric struct {
	Filters        RunaiExternalPodInfoFilters
	ResponseLabels RunaiExternalPodInfoResponseLabels
	Quantile       *float32
}

func (m RunaiExternalPodInfoMetric) GetName() string {
	return "runai_external_pod_info"
}

func (m RunaiExternalPodInfoMetric) GetFilters() interface{} {
	return m.Filters
}

func (m RunaiExternalPodInfoMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m RunaiExternalPodInfoMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiExternalPodInfoMetric) IsAggregation() bool {
	return false
}

func (m RunaiExternalPodInfoMetric) IsQuantile() bool {
	return false
}
