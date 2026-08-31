/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
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
