/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"time"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/metrics"
)

const (
	samplingInterval = 30 * time.Second
	kbToBytes        = 1_000
	mbToBytes        = kbToBytes * 1_000

	metricStoreUsedBytes      = "cache_store_used_bytes"
	metricStoreMaxBytes       = "cache_store_max_bytes"
	metricStoreRecordCount    = "cache_store_record_count"
	metricStoreRecordMaxBytes = "cache_store_record_max_bytes"
)

type MemoryRetrieverer interface {
	RetrieveMemory(storeName string) (int64, error)
}

type LengthRetrieverer interface {
	RetrieveLength(storeName string) (int64, error)
}

type StoreLister interface {
	List() []*options.Store
}

type Sampler struct {
	lengthRetrieverer      LengthRetrieverer
	memoryUsageRetrieverer MemoryRetrieverer
	storeLister            StoreLister
	doneChan               chan struct{}
}

func NewSampler(mr MemoryRetrieverer, lr LengthRetrieverer, storeRegistry StoreLister) *Sampler {
	cacheUsageLabels := []string{
		storeNameLabel,
	}

	metrics.GaugeVec(metricStoreMaxBytes).Register("Store max memory capacity in bytes by store name", cacheUsageLabels)
	metrics.GaugeVec(metricStoreRecordMaxBytes).Register("Store record max size in bytes in cache", cacheUsageLabels)
	metrics.GaugeVec(metricStoreUsedBytes).Register("Store memory usage in bytes by store name", cacheUsageLabels)
	metrics.GaugeVec(metricStoreRecordCount).Register("Store number of entries in cache by store name", cacheUsageLabels)

	return &Sampler{
		memoryUsageRetrieverer: mr,
		lengthRetrieverer:      lr,
		storeLister:            storeRegistry,
		doneChan:               make(chan struct{}, 1),
	}
}

func (s *Sampler) Start() {
	ticker := time.NewTicker(samplingInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.recordMetrics()
			case <-s.doneChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Sampler) recordMetrics() {
	conf := config.Get[*cache_config.Cache]()
	storeConfigs := s.storeLister.List()
	for _, storeConf := range storeConfigs {
		s.recordStoreLimits(storeConf.Name, int64(conf.MaxStoreSizeMB*mbToBytes), int64(conf.MaxRecordSizeKB*kbToBytes))

		if storeMemUsageBytes, err := s.memoryUsageRetrieverer.RetrieveMemory(storeConf.Name); err == nil {
			s.recordStoreUsedBytes(storeConf.Name, storeMemUsageBytes)
		}

		if storeLen, err := s.lengthRetrieverer.RetrieveLength(storeConf.Name); err == nil {
			s.recordStoreRecordsCount(storeConf.Name, storeLen)
		}
	}
}

func (s *Sampler) recordStoreLimits(storeName string, maxUsageBytes int64, maxRecordBytes int64) {
	metrics.GaugeVec(metricStoreMaxBytes).Get().WithLabels(map[string]string{
		storeNameLabel: storeName,
	}).Set(float64(maxUsageBytes))
	metrics.GaugeVec(metricStoreRecordMaxBytes).Get().WithLabels(map[string]string{
		storeNameLabel: storeName,
	}).Set(float64(maxRecordBytes))
}

func (s *Sampler) recordStoreUsedBytes(storeName string, usageBytes int64) {
	metrics.GaugeVec(metricStoreUsedBytes).Get().WithLabels(map[string]string{
		storeNameLabel: storeName,
	}).Set(float64(usageBytes))
}

func (s *Sampler) recordStoreRecordsCount(storeName string, storeLen int64) {
	metrics.GaugeVec(metricStoreRecordCount).Get().WithLabels(map[string]string{
		storeNameLabel: storeName,
	}).Set(float64(storeLen))
}

func (s *Sampler) Stop() {
	close(s.doneChan)
}
