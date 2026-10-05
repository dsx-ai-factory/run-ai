/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/metrics_exporter/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	metrics_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type MetricsCollectorer interface {
	Collect(ctx context.Context, tenantIds []int32) []models.MetricRecord
}

type GaugeExporter struct {
	registry         metrics_pkg.MetricsRegistrar
	exportableMetric *models.ExportableMetric
	collector        MetricsCollectorer
}

func NewGaugeExporter(registry metrics_pkg.MetricsRegistrar, exportableMetric *models.ExportableMetric) *GaugeExporter {
	return &GaugeExporter{
		registry:         registry,
		exportableMetric: exportableMetric,
		collector:        NewMetricsCollector(exportableMetric),
	}
}

func (e *GaugeExporter) Export(ctx context.Context, tenantIds []int32) {
	metricRecords := e.collector.Collect(ctx, tenantIds)
	e.registry.GaugeVec(e.exportableMetric.Name).Get().Reset()
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("exporting %d metric records for metric %s", len(metricRecords), e.exportableMetric.Name)
	for _, metricRecord := range metricRecords {
		e.registry.GaugeVec(e.exportableMetric.Name).Get().WithLabels(metricRecord.Labels).Set(metricRecord.Value)
	}
}
