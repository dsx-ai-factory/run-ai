/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	metrics_pkg "github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	MetricsRegistryName = "workloads_metrics"
)

var exportedMetrics = []*models.ExportableMetric{
	{
		Name:        "external_pod_info",
		Description: "Information about runai pods that does not include any personal data. Pods will be listed for up to 5 minutes after they stop running.",
		LabelsMapping: map[string]string{
			"clusterId":       "cluster_id",
			"node":            "node",
			"pod_uuid":        "pod_id",
			"workload_id":     "workload_id",
			"detailed_status": "status",
		},
		Metric: runaimetrics.RunaiExternalPodInfoMetric{
			Filters: runaimetrics.RunaiExternalPodInfoFilters{
				DetailedStatus: []string{"Error", "CrashLoopBackOff", "StartError"},
			},
		},
	},
}

func InitGoRoutines(ctx context.Context) {
	multiTenancyEnabled := runai_config.Get[*config.MetricsExporter]().MultiTenancyEnabled
	if multiTenancyEnabled == nil || *multiTenancyEnabled {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("multi tenancy enabled is %+v, skipping metrics exporter", multiTenancyEnabled)
		return
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Init go routine for metrics exporter")
	registry := metrics_pkg.Registry(MetricsRegistryName)
	for _, metric := range exportedMetrics {
		registry.GaugeVec(metric.Name).Register(metric.Description, getLabels(metric))
	}
	NewManager(metrics_pkg.Registry(MetricsRegistryName)).Start(ctx, exportedMetrics)
}

func getLabels(metric *models.ExportableMetric) []string {
	labels := []string{tenantIdLabel}
	for _, label := range metric.LabelsMapping {
		labels = append(labels, label)
	}
	return labels
}
