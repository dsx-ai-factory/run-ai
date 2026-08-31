/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMetricsWriter(t *testing.T) {
	writer := newMetricsWriter()
	require.NotNil(t, writer)
	require.NotNil(t, writer.stats)
}

func TestRecordOperation_NoError(t *testing.T) {
	operationCount := mocks.NewMockMetricCounter(t)
	operationCounterVector := mocks.NewMockCounterVector(t)
	operationCounterVector.On("WithLabels", mock.MatchedBy(func(labels map[string]string) bool {
		return labels[errorLabel] == "false"
	})).Return(operationCount).Once()
	operationCount.On("Inc").Return()

	operationDuration := mocks.NewMockMetricHistogram(t)
	operationDurationVector := mocks.NewMockHistogramVector(t)
	operationDurationVector.On("WithLabels", mock.Anything).Return(operationDuration)
	operationDuration.On("Observe", mock.Anything).Return()

	getOperationMissCount := mocks.NewMockMetricCounter(t)
	getOperationMissCountVector := mocks.NewMockCounterVector(t)

	mockStats := &stats{
		operationCount:        operationCounterVector,
		operationDuration:     operationDurationVector,
		getOperationMissCount: getOperationMissCountVector,
	}

	writer := metricsWriter{
		stats: mockStats,
	}
	writer.RecordOperation(string(config.CTRedis), "test-store", OTSet, 1, nil, false)

	operationCount.AssertNumberOfCalls(t, "Inc", 1)
	operationDuration.AssertNumberOfCalls(t, "Observe", 1)
	getOperationMissCount.AssertNumberOfCalls(t, "Inc", 0)
}

func TestRecordOperation_Error(t *testing.T) {
	operationCount := mocks.NewMockMetricCounter(t)
	operationCounterVector := mocks.NewMockCounterVector(t)
	operationCounterVector.On("WithLabels", mock.MatchedBy(func(labels map[string]string) bool {
		return labels[errorLabel] == "true"
	})).Return(operationCount).Once()
	operationCount.On("Inc").Return()

	operationDuration := mocks.NewMockMetricHistogram(t)
	operationDurationVector := mocks.NewMockHistogramVector(t)
	operationDurationVector.On("WithLabels", mock.Anything).Return(operationDuration)
	operationDuration.On("Observe", mock.Anything).Return()

	getOperationMissCount := mocks.NewMockMetricCounter(t)
	getOperationMissCountVector := mocks.NewMockCounterVector(t)

	mockStats := &stats{
		operationCount:        operationCounterVector,
		operationDuration:     operationDurationVector,
		getOperationMissCount: getOperationMissCountVector,
	}

	writer := metricsWriter{
		stats: mockStats,
	}
	writer.RecordOperation(string(config.CTRedis), "test-store", OTSet, 1, errors.New("test error"), false)

	operationCount.AssertNumberOfCalls(t, "Inc", 1)
	operationDuration.AssertNumberOfCalls(t, "Observe", 1)
	getOperationMissCount.AssertNumberOfCalls(t, "Inc", 0)
}

func TestRecordOperation_WithCacheMiss(t *testing.T) {
	operationCount := mocks.NewMockMetricCounter(t)
	operationCounterVector := mocks.NewMockCounterVector(t)
	operationCounterVector.On("WithLabels", mock.MatchedBy(func(labels map[string]string) bool {
		return labels[errorLabel] == "false"
	})).Return(operationCount).Once()
	operationCount.On("Inc").Return()

	operationDuration := mocks.NewMockMetricHistogram(t)
	operationDurationVector := mocks.NewMockHistogramVector(t)
	operationDurationVector.On("WithLabels", mock.Anything).Return(operationDuration)
	operationDuration.On("Observe", mock.Anything).Return()

	getOperationMissCount := mocks.NewMockMetricCounter(t)
	getOperationMissCountVector := mocks.NewMockCounterVector(t)
	getOperationMissCountVector.On("WithLabels", map[string]string{cacheTypeLabel: string(config.CTRedis), storeNameLabel: "test-store"}).Return(getOperationMissCount).Once()
	getOperationMissCount.On("Inc").Return()

	mockStats := &stats{
		operationCount:        operationCounterVector,
		operationDuration:     operationDurationVector,
		getOperationMissCount: getOperationMissCountVector,
	}

	writer := metricsWriter{
		stats: mockStats,
	}
	writer.RecordOperation(string(config.CTRedis), "test-store", OTSet, 1, nil, true)

	operationCount.AssertNumberOfCalls(t, "Inc", 1)
	operationDuration.AssertNumberOfCalls(t, "Observe", 1)
	getOperationMissCount.AssertNumberOfCalls(t, "Inc", 1)
}

func TestRecordOperation_NoError_StoreOperation(t *testing.T) {
	operationCount := mocks.NewMockMetricCounter(t)
	operationCounterVector := mocks.NewMockCounterVector(t)
	operationCounterVector.On("WithLabels", mock.Anything).Return(operationCount)
	operationCount.On("Inc").Return()
	operationCount.On("Add", float64(0)).Return()

	operationDuration := mocks.NewMockMetricHistogram(t)
	operationDurationVector := mocks.NewMockHistogramVector(t)
	operationDurationVector.On("WithLabels", mock.Anything).Return(operationDuration)
	operationDuration.On("Observe", mock.Anything).Return()

	getOperationMissCount := mocks.NewMockMetricCounter(t)
	getOperationMissCountVector := mocks.NewMockCounterVector(t)
	getOperationMissCountVector.On("WithLabels", mock.Anything).Return(getOperationMissCount)
	getOperationMissCount.On("Add", float64(0)).Return()

	mockStats := &stats{
		operationCount:        operationCounterVector,
		operationDuration:     operationDurationVector,
		getOperationMissCount: getOperationMissCountVector,
	}

	writer := metricsWriter{
		stats: mockStats,
	}
	writer.RecordOperation(string(config.CTRedis), "test-store", OTStore, 1, nil, false)

	operationCount.AssertNumberOfCalls(t, "Inc", 1)
	operationDuration.AssertNumberOfCalls(t, "Observe", 1)
	getOperationMissCount.AssertNumberOfCalls(t, "Inc", 0)

	operationCount.AssertNumberOfCalls(t, "Add", 12)
	getOperationMissCount.AssertNumberOfCalls(t, "Add", 1)
}

func TestNewStats(t *testing.T) {
	stats := newStats()
	require.NotNil(t, stats)
	require.NotNil(t, stats.operationCount)
	require.NotNil(t, stats.operationDuration)
}
