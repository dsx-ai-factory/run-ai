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

func GetWorkloadsCountMetric(clusterIds []uuid.UUID, metricGroupBy metricsClient.MetricGroupBy, nodePoolName string) metricsClient.GroupedMetric {
	metric := runaimetrics.RunaiWorkloadsCountMetric{
		Filters: runaimetrics.RunaiWorkloadsCountFilters{
			ClusterId: clusterIds,
		},
		GroupBy: metricGroupBy,
	}
	if nodePoolName != "" {
		metric.Filters.Nodepool = []string{nodePoolName}
	}
	return metric
}
