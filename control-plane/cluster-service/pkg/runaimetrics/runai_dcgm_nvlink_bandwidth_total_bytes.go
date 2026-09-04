/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimetrics

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

type RunaiDcgmNvlinkBandwidthTotalBytesFilters struct {
	ClusterId []uuid.UUID `mapstructure:"clusterId,omitempty"`
	Node      []string    `mapstructure:"node,omitempty"`
	Pod       []string    `mapstructure:"pod_uuid,omitempty"`
	Workload  []string    `mapstructure:"workload_id,omitempty"`
}

type RunaiDcgmNvlinkBandwidthTotalBytesResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
	Node      bool `mapstructure:"node,omitempty"`
	PodId     bool `mapstructure:"pod_uuid,omitempty"`
	Gpu       bool `mapstructure:"gpu,omitempty"`
}

type RunaiDcgmNvlinkBandwidthTotalBytesMetric struct {
	Filters        RunaiDcgmNvlinkBandwidthTotalBytesFilters
	ResponseLabels RunaiDcgmNvlinkBandwidthTotalBytesResponseLabels
	Quantile       *float32
	GroupBy        client.MetricGroupBy
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) GetName() string {
	switch m.GroupBy {
	case client.GroupByGpu:
		return "runai_dcgm_nvlink_bandwidth_total_bytes_per_gpu"
	}

	return "runai_dcgm_nvlink_bandwidth_total_bytes_per_node"
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) GetFilters() any {
	return m.Filters
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) GetResponseLabels() any {
	return m.ResponseLabels
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) IsAggregation() bool {
	return false
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) IsQuantile() bool {
	return false
}

func (m RunaiDcgmNvlinkBandwidthTotalBytesMetric) GetGroupBy() client.MetricGroupBy {
	return m.GroupBy
}
