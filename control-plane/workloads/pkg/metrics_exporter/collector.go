/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"
	"strconv"
	"sync"

	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	tenantIdLabel   = "tenant_id"
	concurrentLimit = 50
)

type MetricsCollector struct {
	concurrentLimitCh chan struct{}
	exportableMetric  *models.ExportableMetric
	clustersClienter  ClustersClienter
}

func NewMetricsCollector(exportableMetric *models.ExportableMetric) *MetricsCollector {
	conf := runai_config.Get[*config.ServicesConfig]()

	return &MetricsCollector{
		concurrentLimitCh: make(chan struct{}, concurrentLimit),
		exportableMetric:  exportableMetric,
		clustersClienter:  newClustersClient(conf.ClusterServiceUrl),
	}
}

func (c *MetricsCollector) Collect(ctx context.Context, tenantIds []int32) []models.MetricRecord {
	resultsChan := make(chan []models.MetricRecord, len(tenantIds))
	var wg sync.WaitGroup

	for _, tenantId := range tenantIds {
		select {
		case <-ctx.Done():
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Operation canceled — stopping metric collection for metric %s", c.exportableMetric.Metric.GetName())
			wg.Wait()
			close(resultsChan)
			return nil
		case c.concurrentLimitCh <- struct{}{}:
			wg.Add(1)
			go func(ctx context.Context, tenantId int32) {
				defer func() {
					wg.Done()
					<-c.concurrentLimitCh
				}()
				records := c.collectTenantMetricRecords(ctx, tenantId)
				if records != nil {
					resultsChan <- records
				}
			}(ctx, tenantId)
		}
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var metricRecords []models.MetricRecord
	for r := range resultsChan {
		metricRecords = append(metricRecords, r...)
	}

	return metricRecords
}

func (c *MetricsCollector) collectTenantMetricRecords(ctx context.Context, tenantId int32) []models.MetricRecord {
	ctx = context.WithValue(ctx, runaictx.CfTenantId, strconv.Itoa(int(tenantId)))
	clusterIds, err := c.clustersClienter.GetClusterIds(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get cluster ids for tenant %d: %v", tenantId, err)
		return nil
	}
	metricRecords := make([]models.MetricRecord, 0)
	for _, clusterId := range clusterIds {
		queryResponse, err := queryMetric(ctx, c.exportableMetric.Metric, tenantId, clusterId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to query metric %s for cluster %s of tenant %d: %v", c.exportableMetric.Metric.GetName(), clusterId, tenantId, err)
			continue
		}

		for _, result := range queryResponse.GetData() {
			metricRecord := c.parseMetricResult(ctx, tenantId, result)
			if metricRecord == nil {
				continue
			}

			metricRecords = append(metricRecords, *metricRecord)
		}
	}

	return metricRecords
}

func (c *MetricsCollector) parseMetricResult(ctx context.Context, tenantId int32, result metrics_api.Result) *models.MetricRecord {
	labels := result.GetMetricLabels()
	if len(labels) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("No labels found for metric %q for tenant %d", c.exportableMetric.Metric.GetName(), tenantId)
		return nil
	}

	measurements := result.GetMeasurements()
	if len(measurements) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("No measurements found for metric %q for tenant %d", c.exportableMetric.Metric.GetName(), tenantId)
		return nil
	}

	mappedLabels := mapLabels(c.exportableMetric.LabelsMapping, labels, c.exportableMetric.Metric.GetName())
	mappedLabels[tenantIdLabel] = strconv.Itoa(int(tenantId))

	value, err := strconv.ParseFloat(measurements[0].GetValue(), 64)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Cannot parse value %q as float for metric %q (defaulting to 0): %v", measurements[0].GetValue(), c.exportableMetric.Metric.GetName(), err)
		value = 0
	}

	return &models.MetricRecord{
		Labels: mappedLabels,
		Value:  value,
	}
}

func queryMetric(ctx context.Context, metric metrics_client.Metric, tenantId int32, clusterId string) (*metrics_api.GetMetrics200Response, error) {
	q, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}
	q.AddFilter("clusterId", clusterId)
	q.SetOverrideTenantId(tenantId)

	res, err := q.Query(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func mapLabels(mapping map[string]string, labels map[string]string, metricName string) map[string]string {
	mappedLabels := make(map[string]string)
	if len(mapping) == 0 || len(labels) == 0 {
		return mappedLabels
	}

	for k, v := range mapping {
		if _, ok := labels[k]; ok {
			mappedLabels[v] = labels[k]
		} else {
			logging.Warnf("Label %q not found in metric response labels metric %s", k, metricName)
		}
	}
	return mappedLabels
}
