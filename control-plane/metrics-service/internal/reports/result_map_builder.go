/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/clusters"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const successStatus = "success"

const metricValueDecimalPrecision = 4

const clusterIdLabel = "clusterId"

type resultMapBuilder struct {
	ClustersClient     ClusterClienter
	ClusterIdToNameMap map[string]string
}

func NewResultMapBuilder(clustersClient *clusters.Client) *resultMapBuilder {
	return &resultMapBuilder{
		ClustersClient:     clustersClient,
		ClusterIdToNameMap: map[string]string{},
	}
}

func (b *resultMapBuilder) Build(results map[string]*api.GetMetrics200Response, labels []string, metricDisplayNames []string, tenantId int32) map[string][]string {
	if len(results) == 0 {
		return map[string][]string{}
	}
	data := b.buildDataMap(metricDisplayNames, results, labels, tenantId)
	return data
}

func (b *resultMapBuilder) buildDataMap(metricDisplayNames []string, results map[string]*api.GetMetrics200Response, labels []string, tenantId int32) map[string][]string {
	numOfMetrics := len(metricDisplayNames)
	data := make(map[string][]string)
	for i, name := range metricDisplayNames {
		err := b.processMetric(results, name, data, labels, numOfMetrics, i, tenantId)
		if err != nil {
			logging.Warnf("Skipping metric %s due to error: %v", name, err)
		}
	}
	return data
}

func (b *resultMapBuilder) processMetric(results map[string]*api.GetMetrics200Response, name string, data map[string][]string, labelNames []string, numOfMetrics int, i int, tenantId int32) error {
	metric := results[name]
	if metric.Status != successStatus {
		return fmt.Errorf("failed status")
	}
	if metric.Data == nil {
		return fmt.Errorf("no data")
	}
	for _, result := range metric.Data {
		err := b.addDataEntry(data, result, labelNames, numOfMetrics, i, tenantId)
		if err != nil {
			logging.Warnf("Skipping metric %s data entry due to error: %v", name, err)
		}
	}
	return nil
}

func (b *resultMapBuilder) addDataEntry(data map[string][]string, metricData api.Result, labelNames []string, numOfMetrics int, index int, tenantId int32) error {
	if metricData.MetricLabels == nil {
		return fmt.Errorf("missing metric labels")
	}
	labelValues, err := b.extractLabelValues(labelNames, metricData.MetricLabels, tenantId)
	if err != nil {
		return fmt.Errorf("missing label values: %v", err)
	}

	if len(metricData.Measurements) == 0 {
		return fmt.Errorf("missing measurement for label set: %s", labelValues)
	}

	if _, ok := data[labelValues]; !ok {
		data[labelValues] = make([]string, numOfMetrics)
	}

	data[labelValues][index] = formatDecimalValue(metricData.Measurements[0].Value)
	return nil
}

func formatDecimalValue(value string) string {
	parsedNumber, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	rounded := round(parsedNumber, metricValueDecimalPrecision)
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func round(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return float64(roundToNearestInt(num*pow)) / pow
}

func roundToNearestInt(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (b *resultMapBuilder) extractLabelValues(neededLabels []string, metricLabels map[string]string, tenantId int32) (string, error) {
	ret := []string{}
	for _, l := range neededLabels {
		val, ok := metricLabels[l]
		if !ok {
			return "", fmt.Errorf("missing label %s", l)
		}

		if l == clusterIdLabel {
			val = b.getClusterLabelValue(val, tenantId)
		}

		ret = append(ret, val)
	}
	return strings.Join(ret, ","), nil
}

func (b *resultMapBuilder) getClusterLabelValue(clusterId string, tenantId int32) string {
	if cachedName, ok := b.ClusterIdToNameMap[clusterId]; ok {
		return cachedName
	}

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	name, err := b.ClustersClient.GetClusterName(ctx, clusterId)
	if err != nil {
		logging.Errorf("Failed to get cluster name for id %s, error: %v", clusterId, err)
		return clusterId
	}
	b.ClusterIdToNameMap[clusterId] = name
	return name
}
