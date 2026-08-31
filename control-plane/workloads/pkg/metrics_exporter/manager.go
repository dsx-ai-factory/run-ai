/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	metrics_pkg "github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	tracing "github.com/run-ai/runai/runai-common-packages/tracing"
)

type Manager struct {
	tenantsClienter        TenantsClienter
	metricsExporterFactory ExporterFactoryer
	metricsScrapeDuration  runaimetrics.MetricHistogram
}

func NewManager(registry metrics_pkg.MetricsRegistrar) *Manager {
	conf := runai_config.Get[*config.ServicesConfig]()
	return &Manager{
		tenantsClienter:        newTenantsClient(conf.TenantsManagerUrl),
		metricsExporterFactory: NewExporterFactory(registry),
		metricsScrapeDuration: metrics_pkg.Histogram("workloads_metrics_scrape_latency_seconds").
			Register("Time taken to scrape workloads metrics in seconds", map[string]string{}, []float64{10, 20, 30}),
	}
}

func (m *Manager) Start(ctx context.Context, metrics []*models.ExportableMetric) {
	go m.start(ctx, metrics)
}

func (m *Manager) start(ctx context.Context, metrics []*models.ExportableMetric) {
	interval := time.Duration(runai_config.Get[*config.MetricsExporter]().IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tracer := tracing.GetTracer()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, span := tracer.StartSpan(ctx, "exportMetrics")

			start := time.Now()
			if err := m.exportMetrics(runCtx, metrics); err != nil {
				logging.WithCtx(runaictx.LoggerHook(runCtx)).Errorf("failed to export metrics: %v", err)
				span.End()
				continue
			}

			if duration := time.Since(start); duration > time.Duration(0.8*float64(interval)) {
				logging.WithCtx(runaictx.LoggerHook(runCtx)).Warnf("metrics scrape duration: %v", duration)
			}
			m.metricsScrapeDuration.Observe(time.Since(start).Seconds())
			span.End()
		}
	}
}

func (m *Manager) exportMetrics(ctx context.Context, metrics []*models.ExportableMetric) error {
	tenantIds, err := m.tenantsClienter.GetTenantIds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant IDs: %w", err)
	}

	for _, metric := range metrics {
		m.exportMetric(ctx, metric, tenantIds)
	}

	return nil
}

func (m *Manager) exportMetric(ctx context.Context, metric *models.ExportableMetric, tenantIds []int32) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("exporting metric %s for tenant IDs %v", metric.Name, tenantIds)
	m.metricsExporterFactory.CreateExporter(metric).Export(ctx, tenantIds)
}
