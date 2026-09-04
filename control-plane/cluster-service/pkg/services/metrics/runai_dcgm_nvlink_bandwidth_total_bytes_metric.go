/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/google/uuid"
)

func GetDcgmNvlinkBandwidthTotalBytesMetric(clusterIds []uuid.UUID, metricGroupBy metricsClient.MetricGroupBy, nodeFilter []string) metricsClient.GroupedMetric {
	metric := runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesMetric{
		Filters: runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesFilters{
			ClusterId: clusterIds,
			Node:      nodeFilter,
		},
		GroupBy: metricGroupBy,
	}
	return metric
}
