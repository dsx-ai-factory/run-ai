/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	agentNamespace = "agent"

	backendSyncMetricName  = "successful_backend_sync_count"
	syncDurationMetricName = "sync_duration"
	syncActionMetricName   = "sync_action"
	syncFailureMetricName  = "sync_failure"

	syncTypeLabel = "sync_type"

	syncHandler = "handler"
	action      = "action"
)

type SyncType string

const (
	Pull SyncType = "Pull"
)

type Action string

const (
	ActionCreateOrUpdate = "create_or_update"
	ActionDelete         = "delete"
)

var getMetricsExporterOnce sync.Once
var initMetricsExporterOnce sync.Once
var metricsExporter *MetricsExporter

type MetricsExporter struct {
	backendSyncCounter *prometheus.CounterVec
	syncDuration       *prometheus.HistogramVec
	syncAction         *prometheus.CounterVec
	syncFailures       *prometheus.CounterVec
}

func GetMetricsExporter() *MetricsExporter {
	getMetricsExporterOnce.Do(func() {
		backendSyncCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: backendSyncMetricName,
			Help: "Counter of successful synchronizations with the backend",
		}, []string{syncTypeLabel})

		syncDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: agentNamespace,
			Name:      syncDurationMetricName,
			Help:      "Duration of resource synchronization operations",
			Buckets:   prometheus.DefBuckets, // TODO adjust default bucketing after reviewing metrics value over 1 week of operation
		}, []string{syncHandler})

		syncAction := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: agentNamespace,
			Name:      syncActionMetricName,
			Help:      "Counter of different synchronizations actions",
		}, []string{syncHandler, action})

		syncFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: agentNamespace,
			Name:      syncFailureMetricName,
			Help:      "Counter of failed synchronizations operations",
		}, []string{syncHandler})

		metricsExporter = &MetricsExporter{
			backendSyncCounter: backendSyncCounter,
			syncDuration:       syncDuration,
			syncFailures:       syncFailures,
			syncAction:         syncAction,
		}
	})

	return metricsExporter
}

func (me *MetricsExporter) Init() {
	initMetricsExporterOnce.Do(func() {
		prometheus.MustRegister(me.backendSyncCounter,
			me.syncDuration,
			me.syncAction,
			me.syncFailures)
		go me.startServer()
	})
}

func (me *MetricsExporter) startServer() {
	http.Handle(config.Config().MetricsPath, promhttp.Handler())

	server := &http.Server{Addr: ":" + config.Config().MetricsPort, Handler: nil}
	if err := server.ListenAndServe(); err != nil {
		log.Error().Msgf("Failed ListenAndServe, err: %v", err.Error())
	}
}

func (me *MetricsExporter) IncBackendSyncCounter(syncType SyncType) {
	me.backendSyncCounter.With(prometheus.Labels{syncTypeLabel: string(syncType)}).Inc()
}

func (me *MetricsExporter) StartSyncTimer(handlerName string) *prometheus.Timer {
	return prometheus.NewTimer(me.syncDuration.WithLabelValues(handlerName))
}

func (me *MetricsExporter) AddResourceAction(handlerName string, action Action, count int) {
	me.syncAction.WithLabelValues(handlerName, string(action)).Add(float64(count))
}

func (me *MetricsExporter) IncSyncFailure(handlerName string) {
	me.backendSyncCounter.WithLabelValues(handlerName).Inc()
}
