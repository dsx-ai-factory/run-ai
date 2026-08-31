/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry"
)

func newMetricsRegistry(promRegistry *prometheus.Registry) *MetricsRegistry {
	return &MetricsRegistry{
		histogramRegistry:    registry.NewHistogramRegistry(promRegistry),
		gaugeRegistry:        registry.NewGaugeRegistry(promRegistry),
		counterRegistry:      registry.NewCounterRegistry(promRegistry),
		summaryRegistry:      registry.NewSummaryRegistry(promRegistry),
		histogramVecRegistry: registry.NewHistogramVecRegistry(promRegistry),
		gaugeVecRegistry:     registry.NewGaugeVecRegistry(promRegistry),
		counterVecRegistry:   registry.NewCounterVecRegistry(promRegistry),
		summaryVecRegistry:   registry.NewSummaryVecRegistry(promRegistry),
		handler:              promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}),
		promRegistry:         promRegistry,
	}
}

// MetricsRegistry is a struct that holds references to different types of metric registries.
// It provides methods to access these registries and retrieve specific metrics.
type MetricsRegistry struct {
	histogramRegistry    registry.HistogramRegisterer
	gaugeRegistry        registry.GaugeRegisterer
	counterRegistry      registry.CounterRegisterer
	summaryRegistry      registry.SummaryRegisterer
	histogramVecRegistry registry.HistogramVecRegisterer
	gaugeVecRegistry     registry.GaugeVecRegisterer
	counterVecRegistry   registry.CounterVecRegisterer
	summaryVecRegistry   registry.SummaryVecRegisterer
	handler              http.Handler
	promRegistry         *prometheus.Registry
}

func (m *MetricsRegistry) Histogram(name string) registry.HistogramAccessor {
	return &registry.HistogramAccess{
		Registry:   m.histogramRegistry,
		MetricName: name,
	}
}

func (m *MetricsRegistry) Gauge(name string) registry.GaugeAccessor {
	return &registry.GaugeAccess{
		Registry:   m.gaugeRegistry,
		MetricName: name,
	}
}

func (m *MetricsRegistry) Counter(name string) registry.CounterAccessor {
	return &registry.CounterAccess{
		Registry:   m.counterRegistry,
		MetricName: name,
	}
}

func (m *MetricsRegistry) Summary(name string) registry.SummaryAccessor {
	return &registry.SummaryAccess{
		Registry:   m.summaryRegistry,
		MetricName: name,
	}
}

func (m *MetricsRegistry) Handler() http.Handler {
	return m.handler
}

func (m *MetricsRegistry) HistogramVec(name string) registry.HistogramVecAccessor {
	return &registry.HistogramVecAccess{
		Registry: m.histogramVecRegistry,
		VecName:  name,
	}
}

func (m *MetricsRegistry) GaugeVec(name string) registry.GaugeVecAccessor {
	return &registry.GaugeVecAccess{
		Registry: m.gaugeVecRegistry,
		VecName:  name,
	}
}

func (m *MetricsRegistry) CounterVec(name string) registry.CounterVecAccessor {
	return &registry.CounterVecAccess{
		Registry: m.counterVecRegistry,
		VecName:  name,
	}
}

func (m *MetricsRegistry) SummaryVec(name string) registry.SummaryVecAccessor {
	return &registry.SummaryVecAccess{
		Registry: m.summaryVecRegistry,
		VecName:  name,
	}
}

func (m *MetricsRegistry) GetPrometheusRegistry() *prometheus.Registry {
	return m.promRegistry
}

func (m *MetricsRegistry) BuildMetricsString() (string, error) {
	metrics, _ := m.promRegistry.Gather()

	var formattedMetrics bytes.Buffer
	enc := expfmt.NewEncoder(&formattedMetrics, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, metric := range metrics {
		if err := enc.Encode(metric); err != nil {
			return "", fmt.Errorf("encoding gathered metrics failed: %w", err)
		}
	}

	return formattedMetrics.String(), nil
}
