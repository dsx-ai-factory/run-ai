/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	ErrorMetricName = "errors_count"

	errorCounterReasonLabel = "reason"
)

var getMetricsExporterOnce sync.Once
var initMetricsExporterOnce sync.Once
var metricsExporter *MetricsExporter

type MetricsExporter struct {
	errorsCounter *prometheus.CounterVec
	// the handler for the backendSyncCounter
	promhttpHandler http.Handler
}

func GetMetricsExporter() *MetricsExporter {
	getMetricsExporterOnce.Do(func() {
		errorsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: ErrorMetricName,
			Help: "Counter of errors while handling CRs",
		}, []string{errorCounterReasonLabel})
		metricsExporter = &MetricsExporter{
			errorsCounter: errorsCounter,
		}
	})

	return metricsExporter
}

func (ex *MetricsExporter) Init() {
	initMetricsExporterOnce.Do(func() {
		prometheus.MustRegister(ex.errorsCounter)
		go ex.startServer()
	})
}

func (ex *MetricsExporter) startServer() {
	ex.promhttpHandler = promhttp.Handler()

	http.Handle(config.Config().MetricsPath, ex)

	server := &http.Server{Addr: ":" + config.Config().MetricsPort, Handler: nil}
	if err := server.ListenAndServe(); err != nil {
		zap.S().Errorf("Failed ListenAndServe in MetricsExporter, err: %v", err.Error())
	}
}

func (ex *MetricsExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ex.promhttpHandler.ServeHTTP(w, r)
}

func (ex *MetricsExporter) AddErrorCounter(err error) {
	reason := errors.ReasonForError(err)
	ex.errorsCounter.With(prometheus.Labels{errorCounterReasonLabel: string(reason)}).Inc()
}
