/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
)

type ClusterVersionsHistoryGetter interface {
	GetClustersVersionsHistories(ctx context.Context, clusterIds []string) (map[string]*clusters.ClusterVersionHistory, error)
}
type QueryVersionPicker interface {
	Pick(ctx context.Context, cvMap map[string]*clusters.ClusterVersionHistory, r *services.MetricInformationRange) string
}

type QueryInjector interface {
	Inject(query string, metricInformation *services.MetricInformation) (string, error)
}

type MetricConfigurationProcessor interface {
	Process(metricConfigs map[string]*MetricConfiguration) (map[string]*ProcessedMetricConfiguration, error)
}

type MetricQueryPicker interface {
	GetQueryByVersion(metricName string, version string) (string, error)
	GetQueryMinVersion(metricName string) (*version.Version, error)
}
