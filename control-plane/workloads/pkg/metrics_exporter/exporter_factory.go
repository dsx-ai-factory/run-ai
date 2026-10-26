/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/metrics_exporter/models"
	metrics_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
)

type MetricsExporterer interface {
	Export(ctx context.Context, tenantIds []int32)
}

type ExporterFactoryer interface {
	CreateExporter(metric *models.ExportableMetric) MetricsExporterer
}

type ExporterFactory struct {
	registry metrics_pkg.MetricsRegistrar
}

func NewExporterFactory(registry metrics_pkg.MetricsRegistrar) *ExporterFactory {
	return &ExporterFactory{
		registry: registry,
	}
}

func (e *ExporterFactory) CreateExporter(metric *models.ExportableMetric) MetricsExporterer {
	return NewGaugeExporter(e.registry, metric)
}
