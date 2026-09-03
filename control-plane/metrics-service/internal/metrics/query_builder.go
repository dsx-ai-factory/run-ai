/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/hashicorp/go-version"
)

type PromMetricQueryBuilder struct {
	clusterVersionsHistoryGetter ClusterVersionsHistoryGetter
	queryVersionPicker           QueryVersionPicker
	metricQueryPicker            MetricQueryPicker
	injectorChain                QueryInjector
}

func NewPromMetricQueryBuilder(clusterVersionsHistoryGetter ClusterVersionsHistoryGetter, queryVersionPicker QueryVersionPicker,
	metricQueryPicker MetricQueryPicker, injectorChain QueryInjector,
) *PromMetricQueryBuilder {
	return &PromMetricQueryBuilder{
		clusterVersionsHistoryGetter: clusterVersionsHistoryGetter,
		queryVersionPicker:           queryVersionPicker,
		metricQueryPicker:            metricQueryPicker,
		injectorChain:                injectorChain,
	}
}

func (pmqb *PromMetricQueryBuilder) Build(ctx context.Context, metricInformation *services.MetricInformation) (string, error) {
	versionHistories, err := pmqb.clusterVersionsHistoryGetter.GetClustersVersionsHistories(ctx, metricInformation.Filters["clusterId"])
	if err != nil {
		return "", err
	}

	minVersion, err := pmqb.metricQueryPicker.GetQueryMinVersion(metricInformation.Name)
	if err != nil {
		return "", err
	}

	if metricVersionExceedClustersVersions(versionHistories, minVersion) {
		return "", fmt.Errorf("%w %s", errors.ErrClusterVersionNotSupported, metricInformation.Name)
	}

	version := pmqb.queryVersionPicker.Pick(ctx, versionHistories, metricInformation.Range)

	query, err := pmqb.metricQueryPicker.GetQueryByVersion(metricInformation.Name, version)
	if err != nil {
		return "", err
	}

	injectedQuery, err := pmqb.injectorChain.Inject(query, metricInformation)
	if err != nil {
		return "", err
	}

	return injectedQuery, nil
}

func metricVersionExceedClustersVersions(cvMap map[string]*clusters.ClusterVersionHistory, minVersion *version.Version) bool {
	// if no cluster versions found, we assume that one of the clusters might be in the correct version
	if len(cvMap) == 0 {
		return false
	}
	for _, cv := range cvMap {
		if cv.CurrentVersion == nil || cv.CurrentVersion.GreaterThanOrEqual(minVersion) {
			return false
		}
	}

	return true
}
