/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus_monitor

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/sdk/reports"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	"github.com/prometheus/common/model"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type performanceMetricPair struct {
	typ scale.MetricType
	val reports.PerformanceMetric
}

type prometheusPerformanceMetricExtractorFn func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error)

var prometheusPerformanceMetricExtractors = []prometheusPerformanceMetricExtractorFn{
	func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error) {
		query := fmt.Sprintf(`increase(prometheus_target_scrapes_exceeded_sample_limit_total{namespace="runai",container="prometheus"}[%ds])`, int(time.Since(since).Seconds()))
		val, err := extractUniqueInstantVectorValue(ctx, promClient, query)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", scale.MetricTypeTargetScrapesExceededSampleLimitTotal, err)
		}
		return &performanceMetricPair{
			typ: scale.MetricTypeTargetScrapesExceededSampleLimitTotal,
			val: val,
		}, nil
	},
	func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error) {
		query := fmt.Sprintf(`deriv(prometheus_remote_storage_samples_pending[%ds])`, int(time.Since(since).Seconds()))
		val, err := extractUniqueInstantVectorValue(ctx, promClient, query)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", scale.MetricTypeRemoteWriteQueueBacklogIncreaseRate, err)
		}
		return &performanceMetricPair{
			typ: scale.MetricTypeRemoteWriteQueueBacklogIncreaseRate,
			val: val,
		}, nil
	},
	func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error) {
		query := fmt.Sprintf(`increase(prometheus_remote_storage_samples_failed_total{namespace="runai",container="prometheus"}[%ds])`, int(time.Since(since).Seconds()))
		val, err := extractUniqueInstantVectorValue(ctx, promClient, query)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", scale.MetricTypeRemoteWriteSamplesFailedTotal, err)
		}
		return &performanceMetricPair{
			typ: scale.MetricTypeRemoteWriteSamplesFailedTotal,
			val: val,
		}, nil
	},
	func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error) {
		query := fmt.Sprintf(`increase(kube_pod_container_status_restarts_total{namespace="runai", container="prometheus"}[%ds])`, int(time.Since(since).Seconds()))
		val, err := extractUniqueInstantVectorValue(ctx, promClient, query)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", scale.MetricTypeCrashRate, err)
		}
		return &performanceMetricPair{
			typ: scale.MetricTypeCrashRate,
			val: val,
		}, nil
	},
	func(ctx context.Context, promClient v1.API, since time.Time) (*performanceMetricPair, error) {
		query := fmt.Sprintf(`increase(prometheus_target_sync_failed_total{namespace="runai",container="prometheus"}[%ds])`, int(time.Since(since).Seconds()))

		queryResult, _, _, err := promClient.Query(ctx, query, time.Now())
		if err != nil {
			return nil, fmt.Errorf("failed querying %s: %w", query, err)
		}

		vec, ok := queryResult.(model.Vector)
		if !ok {
			return nil, fmt.Errorf("failed converting to a vector: %w", err)
		}

		scrapeJobToTargetRegex := regexp.MustCompile(`serviceMonitor/runai/([^/]+)/0`)
		targetSyncFailureCount := map[string]float64{}
		for _, sample := range vec {
			scrapeJobLabel := sample.Metric["scrape_job"]
			if !scrapeJobLabel.IsValid() {
				return nil, fmt.Errorf("scrape_job label is not valid")
			}

			// Extract the target name from the label by applying "serviceMonitor/runai/([^/]+)/0" regex
			matches := scrapeJobToTargetRegex.FindStringSubmatch(string(scrapeJobLabel))
			if len(matches) != 2 {
				// Irrelevant target
				continue
			}
			targetName := matches[1]
			targetSyncFailureCount[targetName] += float64(sample.Value)
		}

		return &performanceMetricPair{
			typ: scale.MetricTypeTargetSyncFailedTotal,
			val: targetSyncFailureCount,
		}, nil
	},
}

func extractUniqueInstantVectorValue(ctx context.Context, promClient v1.API, query string) (*model.SampleValue, error) {
	queryResult, _, _, err := promClient.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed querying %s: %w", query, err)
	}

	vec, ok := queryResult.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("failed converting to a vector: %w", err)
	}

	if len(vec) != 1 {
		return nil, fmt.Errorf("unexpected vector %s (query %s) length: %d", vec, query, len(vec))
	}

	return &vec[0].Value, nil
}
