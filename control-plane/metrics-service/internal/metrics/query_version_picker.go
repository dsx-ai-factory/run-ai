/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/clusters"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type queryVersionPicker struct{}

func NewQueryVersionPicker() *queryVersionPicker {
	return &queryVersionPicker{}
}

func (c *queryVersionPicker) Pick(ctx context.Context, cvMap map[string]*clusters.ClusterVersionHistory, r *services.MetricInformationRange) string {
	if len(cvMap) == 0 {
		return CROSS_CLUSTER_VERSION
	}
	start, end := getStartEndTimes(r)
	version := ""
	for _, cv := range cvMap {
		res := findClusterVersion(*cv, start, end)
		if res == clusters.DevVersion.String() || (version != res && version != "") {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("multiple clusters, going to use cross version")
			return CROSS_CLUSTER_VERSION
		}

		version = res
	}

	return version
}

func getStartEndTimes(r *services.MetricInformationRange) (*time.Time, *time.Time) {
	if r == nil {
		return nil, nil
	}
	return &r.Start, &r.End
}

func findClusterVersion(cv clusters.ClusterVersionHistory, start, end *time.Time) string {
	if len(cv.Versions) == 0 {
		return CROSS_CLUSTER_VERSION
	}
	if start == nil || end == nil {
		return currentClusterVersion(cv)
	}
	if start.After(*end) {
		return CROSS_CLUSTER_VERSION
	}

	return findInclusiveTimeRange(cv, start, end)
}

func currentClusterVersion(cv clusters.ClusterVersionHistory) string {
	if cv.CurrentVersion == nil {
		return CROSS_CLUSTER_VERSION
	}
	return cv.CurrentVersion.String()
}

func findInclusiveTimeRange(cv clusters.ClusterVersionHistory, start, end *time.Time) string {
	for _, v := range cv.Versions {
		if v.Start.Before(*start) && v.End.After(*end) {
			return v.Version.String()
		}
	}
	return CROSS_CLUSTER_VERSION
}
