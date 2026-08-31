/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/clock"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"go.uber.org/zap"
)

type SyncType string

const (
	ClusterInfoPush SyncType = "ClusterInfoPush"

	MetricName = "successful_backend_sync_count"

	syncTypeLabel           = "sync_type"
	errorMetricName         = "errors_count"
	cpExecutionDurationName = "control_plane_execution_duration_seconds"
	errorCounterReasonLabel = "reason"
	resourceLabel           = "resource"
	operationLabel          = "operation"
)

var getMetricsExporterOnce sync.Once
var initMetricsExporterOnce sync.Once
var metricsExporter *MetricsExporter

type MetricsExporter struct {
	backendSyncCounter  *prometheus.CounterVec
	errorsCounter       *prometheus.CounterVec
	cpExecutionDuration *prometheus.HistogramVec
	// the handler for the backendSyncCounter
	promHttpHandler http.Handler
	clock           clock.Clock
}

func GetMetricsExporter() *MetricsExporter {
	getMetricsExporterOnce.Do(func() {
		backendSyncCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: MetricName,
			Help: "Counter of successful synchronizations with the backend",
		}, []string{syncTypeLabel})
		errorsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: errorMetricName,
			Help: "Counter of errors",
		}, []string{errorCounterReasonLabel, resourceLabel})
		cpExecutionDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    cpExecutionDurationName,
			Help:    "Duration of control plane data execution in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{resourceLabel, operationLabel})
		metricsExporter = &MetricsExporter{
			backendSyncCounter:  backendSyncCounter,
			errorsCounter:       errorsCounter,
			cpExecutionDuration: cpExecutionDuration,
			clock:               clock.RealClock{},
		}
	})

	return metricsExporter
}

//nolint:unused
func (ex *MetricsExporter) _sinceInSeconds(start time.Time) float64 {
	return ex.clock.Since(start).Seconds()
}

func (ex *MetricsExporter) Init() {
	initMetricsExporterOnce.Do(func() {
		prometheus.MustRegister(ex.errorsCounter)
		prometheus.MustRegister(ex.backendSyncCounter)
		prometheus.MustRegister(ex.cpExecutionDuration)
		go ex.startServer()
	})
}

func (ex *MetricsExporter) startServer() {
	ex.promHttpHandler = promhttp.Handler()

	http.Handle(config.Config().MetricsPath, ex)

	server := &http.Server{Addr: ":" + config.Config().MetricsPort, Handler: nil}
	if err := server.ListenAndServe(); err != nil {
		zap.S().Errorf("Failed ListenAndServe in MetricsExporter, err: %v", err.Error())
	}
}

// ServeHTTP - to handle both handlers for one metrics path
func (ex *MetricsExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ex.promHttpHandler.ServeHTTP(w, r)
}

func (ex *MetricsExporter) IncBackendSyncCounter(syncType SyncType) {
	ex.backendSyncCounter.With(prometheus.Labels{syncTypeLabel: string(syncType)}).Inc()
}

func (ex *MetricsExporter) AddErrorCounter(err error, resource constants.ClusterResource) {
	reason := errors.ReasonForError(err)
	ex.errorsCounter.With(
		prometheus.Labels{
			errorCounterReasonLabel: string(reason),
			resourceLabel:           string(resource)}).Inc()
}

func (ex *MetricsExporter) MeasureExecutionTime(resource constants.ClusterResource, operation constants.OperationName) func() {
	start := ex.clock.Now()
	return func() {
		duration := time.Since(start).Seconds()
		ex.cpExecutionDuration.With(
			prometheus.Labels{
				resourceLabel:  string(resource),
				operationLabel: string(operation)},
		).Observe(duration)
	}
}
