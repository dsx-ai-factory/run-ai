/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"errors"
	"testing"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/observability/mocks"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	metrics_testing "github.com/run-ai/runai/runai-common-packages/metrics/testing"
)

func TestSampler_RecordMetrics_Success(t *testing.T) {
	conf := setupCacheConfig()
	memRetriever := mocks.NewMockMemoryRetrieverer(t)
	lenRetriever := mocks.NewMockLengthRetrieverer(t)
	storeConfig := &options.Store{Name: "store", Options: &options.StoreOptions{MaxStoreSizeMB: 100, MaxRecordSizeKB: conf.MaxRecordSizeKB}}
	usedBytes, recordCount := int64(5000000), int64(100)

	manager := metrics_testing.SetMock(t)

	vecAccessorMaxBytes := manager.GaugeVec(metricStoreMaxBytes)
	vecAccessorMaxBytes.Get().WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxStoreSizeMB * 1_000_000))

	manager.GaugeVec(metricStoreRecordMaxBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxRecordSizeKB * 1_000))

	manager.GaugeVec(metricStoreUsedBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(usedBytes))

	manager.GaugeVec(metricStoreRecordCount).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(recordCount))

	memRetriever.On("RetrieveMemory", "store").Return(usedBytes, nil).Once()
	lenRetriever.On("RetrieveLength", "store").Return(recordCount, nil).Once()

	sampler := &Sampler{
		memoryUsageRetrieverer: memRetriever,
		lengthRetrieverer:      lenRetriever,
		storeLister:            initStoreLister(t, []*options.Store{storeConfig}),
		doneChan:               make(chan struct{}, 1),
	}
	sampler.recordMetrics()

	memRetriever.AssertExpectations(t)
	lenRetriever.AssertExpectations(t)
}

func TestSampler_RecordMetrics_MemoryRetrievalError(t *testing.T) {
	conf := setupCacheConfig()

	memRetriever := mocks.NewMockMemoryRetrieverer(t)
	lenRetriever := mocks.NewMockLengthRetrieverer(t)
	storeConfig := &options.Store{Name: "store", Options: &options.StoreOptions{MaxStoreSizeMB: 100, MaxRecordSizeKB: conf.MaxRecordSizeKB}}

	manager := metrics_testing.SetMock(t)

	manager.GaugeVec(metricStoreMaxBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxStoreSizeMB * 1_000_000))

	manager.GaugeVec(metricStoreRecordMaxBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxRecordSizeKB * 1_000))

	manager.GaugeVec(metricStoreRecordCount).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(100))

	memRetriever.On("RetrieveMemory", "store").Return(int64(-1), errors.New("memory retrieval failed")).Once()
	lenRetriever.On("RetrieveLength", "store").Return(int64(100), nil).Once()

	sampler := &Sampler{
		memoryUsageRetrieverer: memRetriever,
		lengthRetrieverer:      lenRetriever,
		storeLister:            initStoreLister(t, []*options.Store{storeConfig}),
		doneChan:               make(chan struct{}, 1),
	}
	sampler.recordMetrics()

	memRetriever.AssertExpectations(t)
	lenRetriever.AssertExpectations(t)
}

func TestSampler_RecordMetrics_LengthRetrievalError(t *testing.T) {
	conf := setupCacheConfig()

	memRetriever := mocks.NewMockMemoryRetrieverer(t)
	lenRetriever := mocks.NewMockLengthRetrieverer(t)
	storeConfig := &options.Store{
		Name: "store1", Options: &options.StoreOptions{MaxStoreSizeMB: 100, MaxRecordSizeKB: conf.MaxRecordSizeKB},
	}

	memRetriever.On("RetrieveMemory", "store1").Return(int64(5000000), nil).Once()
	lenRetriever.On("RetrieveLength", "store1").Return(int64(0), errors.New("length retrieval failed")).Once()

	manager := metrics_testing.SetMock(t)

	manager.GaugeVec(metricStoreMaxBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxStoreSizeMB * 1_000_000))

	manager.GaugeVec(metricStoreRecordMaxBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(storeConfig.Options.MaxRecordSizeKB * 1_000))

	manager.GaugeVec(metricStoreUsedBytes).Get().
		WithLabels(map[string]string{storeNameLabel: storeConfig.Name}).Set(float64(5000000))

	sampler := &Sampler{
		memoryUsageRetrieverer: memRetriever,
		lengthRetrieverer:      lenRetriever,
		storeLister:            initStoreLister(t, []*options.Store{storeConfig}),
		doneChan:               make(chan struct{}, 1),
	}
	sampler.recordMetrics()

	memRetriever.AssertExpectations(t)
	lenRetriever.AssertExpectations(t)
}

func setupCacheConfig() *cache_config.Cache {
	runai_config.Init(nil)
	runai_config.Set(&cache_config.Cache{
		MaxStoreSizeMB:  100,
		MaxRecordSizeKB: 1000,
		ServiceName:     "test-service",
	})
	return runai_config.Get[*cache_config.Cache]()
}

func initStoreLister(t *testing.T, stores []*options.Store) StoreLister {
	srm := mocks.NewMockStoreLister(t)
	srm.On("List").Return(stores)
	return srm
}
