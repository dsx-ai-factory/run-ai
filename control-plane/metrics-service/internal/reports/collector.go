/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"k8s.io/utils/ptr"
)

type Collector struct {
	ReportsUpdater   ReportsUpdater
	MetricQuerier    services.MetricQuerier
	S3Saver          S3Saver
	ResultMapBuilder ResultMapBuilder
}

func NewCollector(metricQueryBuilder services.MetricQueryBuilder, metricQuerier services.MetricQuerier, clustersClient *clusters.Client) (*Collector, error) {
	csvs3Saver, err := NewCSVS3Saver()
	if err != nil {
		return nil, err
	}
	return &Collector{
		ReportsUpdater:   dao.NewPostgresReportsDao(),
		MetricQuerier:    metricQuerier,
		S3Saver:          csvs3Saver,
		ResultMapBuilder: NewResultMapBuilder(clustersClient),
	}, nil
}

func (c *Collector) Collect(report models.Report) {
	err := c.ReportsUpdater.UpdateReportStatus(report.TenantId, report.ID.String(), api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_PENDING))
	if err != nil {
		return
	}

	res, err := c.queryMetrics(report)
	if err != nil {
		logging.Errorf("Failed to query metric of the report, tenant_id [%d] report_id: [%s]. err: %v", report.TenantId, report.ID.String(), err)
		_ = c.ReportsUpdater.UpdateReportStatusWithMessage(report.TenantId, report.ID.String(), api.REPORTSTATUS_FAILED, err.Error(), nil)
		return
	}
	displayMetricsNames := buildMetricDisplayNamesFromReport(report)
	displayLabelsNames := buildLabelDisplayNames(report.Labels)
	labels := buildLabelNames(report.Labels)
	data := c.ResultMapBuilder.Build(res, labels, displayMetricsNames, report.TenantId)
	iterator := iterateSortedResultMap(data)
	err = c.S3Saver.Save(report.TenantId, report.ID.String(), iterator, displayLabelsNames, displayMetricsNames)
	if err != nil {
		logging.Errorf("Failed to save the report, tenant_id [%d] report_id: [%s]. err: %v", report.TenantId, report.ID.String(), err)
		_ = c.ReportsUpdater.UpdateReportStatusWithMessage(report.TenantId, report.ID.String(), api.REPORTSTATUS_FAILED, err.Error(), nil)
		return
	}

	_ = c.ReportsUpdater.UpdateReportStatus(report.TenantId, report.ID.String(), api.REPORTSTATUS_READY, nil)
}

func (c *Collector) queryMetrics(report models.Report) (map[string]*api.GetMetrics200Response, error) {
	finalData := map[string]*api.GetMetrics200Response{}
	for _, metric := range report.Metrics {
		res, err := c.queryMetric(report, metric.Query, metric.DisplayName)
		if err != nil {
			return nil, err
		}

		finalData[metric.DisplayName] = res
	}

	return finalData, nil
}

func (c *Collector) queryMetric(report models.Report, query string, metricDisplayName string) (*api.GetMetrics200Response, error) {
	res, err := c.MetricQuerier.Query(context.Background(), report.TenantId, query, &services.MetricInformationRange{
		Start:       report.Start,
		End:         report.End,
		Aggregation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query metric %s err %v", metricDisplayName, err)
	}

	return res, nil
}

func buildMetricDisplayNamesFromReport(report models.Report) []string {
	ret := make([]string, len(report.Metrics))
	for i, metric := range report.Metrics {
		ret[i] = metric.DisplayName
	}
	return ret
}

func buildLabelDisplayNames(labelArray models.LabelArray) []string {
	ret := make([]string, len(labelArray))
	for i, l := range labelArray {
		ret[i] = l.DisplayName
	}
	return ret
}

func buildLabelNames(labelArray models.LabelArray) []string {
	ret := make([]string, len(labelArray))
	for i, l := range labelArray {
		ret[i] = l.Name
	}
	return ret
}
