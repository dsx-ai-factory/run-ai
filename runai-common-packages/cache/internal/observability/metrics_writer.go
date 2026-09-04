/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"strconv"
	"time"

	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/runaimetrics"
)

const (
	operationLabel = "operation"
	cacheTypeLabel = "cache_type"
	errorLabel     = "error"
	storeNameLabel = "store_name"
)

var (
	cacheOperationLabels = []string{
		operationLabel,
		cacheTypeLabel,
		storeNameLabel,
		errorLabel,
	}
	cacheMissCountLabels = []string{cacheTypeLabel, storeNameLabel}

	operationTypeLabelOptions = []OperationType{
		OTGet,
		OTSet,
		OTDel,
		OTStore,
		OTClearStore,
		OTClearCache,
	}
	errorLabelOptions = []bool{true, false}
)

type metricsWriter struct {
	stats *stats
}

func newMetricsWriter() *metricsWriter {
	return &metricsWriter{
		stats: newStats(),
	}
}

func (mw *metricsWriter) RecordOperation(cacheType string, storeName string, operationType OperationType, duration time.Duration, err error, miss bool) {
	if isCreateStoreOperation(operationType) {
		mw.stats.initializeMetrics(storeName)
	}

	labels := map[string]string{
		operationLabel: string(operationType),
		cacheTypeLabel: cacheType,
		storeNameLabel: storeName,
		errorLabel:     strconv.FormatBool(err != nil),
	}

	mw.stats.operationCount.WithLabels(labels).Inc()
	mw.stats.operationDuration.WithLabels(labels).Observe(duration.Seconds())

	if miss {
		mw.stats.getOperationMissCount.WithLabels(map[string]string{cacheTypeLabel: cacheType, storeNameLabel: storeName}).Inc()
	}
}

type stats struct {
	operationCount        runaimetrics.CounterVector
	operationDuration     runaimetrics.HistogramVector
	getOperationMissCount runaimetrics.CounterVector
}

func newStats() *stats {
	s := &stats{
		operationCount:        metrics.CounterVec("cache_operation_count").Register("Number of operations executed", cacheOperationLabels),
		operationDuration:     metrics.HistogramVec("cache_operation_duration").Register("Time taken to execute an operation", cacheOperationLabels, []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1}),
		getOperationMissCount: metrics.CounterVec("cache_get_operation_miss_count").Register("Number of cache get operations that resulted in a miss", cacheMissCountLabels),
	}
	return s
}

// Initializes all metric label combinations to avoid missing time-series in Prometheus when the service resets.
func (s *stats) initializeMetrics(storeName string) {
	cacheType := string(config.Get[*cache_config.Cache]().Type)

	initializeMetric(s.operationCount, getCacheOperationLabelCombos(cacheType, storeName))
	initializeMetric(s.getOperationMissCount, getCacheMissCountLabelCombos(cacheType, storeName))
}

func initializeMetric(metric runaimetrics.CounterVector, labelCombos []map[string]string) {
	for _, labels := range labelCombos {
		metric.WithLabels(labels).Add(0)
	}
}

func getCacheOperationLabelCombos(cacheType string, storeName string) []map[string]string {
	var combos []map[string]string
	for _, op := range operationTypeLabelOptions {
		for _, err := range errorLabelOptions {
			combos = append(combos, map[string]string{
				operationLabel: string(op),
				cacheTypeLabel: cacheType,
				storeNameLabel: storeName,
				errorLabel:     strconv.FormatBool(err),
			})
		}
	}
	return combos
}

func getCacheMissCountLabelCombos(cacheType string, storeName string) []map[string]string {
	return []map[string]string{
		{cacheTypeLabel: cacheType, storeNameLabel: storeName},
	}
}

func isCreateStoreOperation(operationType OperationType) bool {
	return operationType == OTStore
}
