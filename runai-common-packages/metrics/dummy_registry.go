/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/registry"
	"github.com/prometheus/client_golang/prometheus"
)

type DummyMetricsRegistrar struct{}

func (d *DummyMetricsRegistrar) Histogram(name string) registry.HistogramAccessor {
	return &registry.DummyHistogramAccessor{}
}

func (d *DummyMetricsRegistrar) Counter(name string) registry.CounterAccessor {
	return &registry.DummyCounterAccessor{}
}

func (d *DummyMetricsRegistrar) Gauge(name string) registry.GaugeAccessor {
	return &registry.DummyGaugeAccessor{}
}

func (d *DummyMetricsRegistrar) Summary(name string) registry.SummaryAccessor {
	return &registry.DummySummaryAccessor{}
}

func (d *DummyMetricsRegistrar) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (d *DummyMetricsRegistrar) HistogramVec(name string) registry.HistogramVecAccessor {
	return &registry.DummyHistogramVecAccessor{}
}

func (d *DummyMetricsRegistrar) CounterVec(name string) registry.CounterVecAccessor {
	return &registry.DummyCounterVecAccessor{}
}

func (d *DummyMetricsRegistrar) GaugeVec(name string) registry.GaugeVecAccessor {
	return &registry.DummyGaugeVecAccessor{}
}

func (d *DummyMetricsRegistrar) SummaryVec(name string) registry.SummaryVecAccessor {
	return &registry.DummySummaryVecAccessor{}
}

func (d *DummyMetricsRegistrar) GetPrometheusRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func (d *DummyMetricsRegistrar) BuildMetricsString() (string, error) {
	return "", nil
}
