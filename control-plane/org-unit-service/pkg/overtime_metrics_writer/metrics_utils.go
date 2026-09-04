/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package overtime_metrics_writer

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	clusterIdLabel      = "clusterId"
	departmentNameLabel = "department"
	projectNameLabel    = "project"
)

func getOvertimeMetricResults(ctx context.Context, tenantId string, overtimeMetricRange dao.OvertimeMetricRange,
	overtimeMetricType OvertimeMetricType, isProject bool,
) ([]metrics_api.Result, error) {
	metric, err := getOvertimeMetric(overtimeMetricType, isProject)
	if err != nil {
		return nil, err
	}

	return queryMetric(ctx, tenantId, metric, overtimeMetricRange)
}

func getOvertimeMetric(overtimeMetricType OvertimeMetricType, isProject bool) (metrics_client.Metric, error) {
	var metric metrics_client.Metric
	switch overtimeMetricType {
	case GpuAllocationAvg:
		if isProject {
			metric = runaimetrics.RunaiAllocatedGpuCountAvgPerProjectMetric{}
		} else {
			metric = runaimetrics.RunaiAllocatedGpuCountAvgPerDepartmentMetric{}
		}
	case GpuUtilizationAvg:
		if isProject {
			metric = runaimetrics.RunaiGpuUtilizationAvgPerProjectMetric{}
		} else {
			metric = runaimetrics.RunaiGpuUtilizationAvgPerDepartmentMetric{}
		}
	case GpuMemoryUtilizationAvg:
		if isProject {
			metric = runaimetrics.RunaiGpuMemoryUtilizationAvgPerProjectMetric{}
		} else {
			metric = runaimetrics.RunaiGpuMemoryUtilizationAvgPerDepartmentMetric{}
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", overtimeMetricType)
	}

	return metric, nil
}

func queryMetric(ctx context.Context, tenantId string, metric metrics_client.Metric, overtimeMetricRange dao.OvertimeMetricRange) ([]metrics_api.Result, error) {
	query, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}

	tenantUid, err := strconv.ParseInt(tenantId, 10, 64)
	if err != nil {
		return nil, err
	}

	end := time.Now().UTC()
	var start time.Time

	switch overtimeMetricRange {
	case dao.RangeDay:
		start = end.AddDate(0, 0, -1)
	case dao.RangeWeek:
		start = end.AddDate(0, 0, -7)
	case dao.RangeMonth:
		start = end.AddDate(0, -1, 0)
	}

	query.SetOverrideTenantId(int32(tenantUid))
	query.SetRange(start, end)

	res, err := query.Query(ctx)
	return res.GetData(), err
}

func parseTenantOvertimeData(ctx context.Context, queryResults []metrics_api.Result,
	overtimeMetricRange dao.OvertimeMetricRange, overtimeMetricType OvertimeMetricType,
	queryProjects bool, result map[string]map[string]*models.OvertimeData,
) {
	orgUnitNameLabel := projectNameLabel
	if !queryProjects {
		orgUnitNameLabel = departmentNameLabel
	}

	if len(queryResults) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("No results for query %v / %v for %ss",
			overtimeMetricRange, overtimeMetricType, orgUnitNameLabel)
		return
	}

	for _, queryResult := range queryResults {
		parseOrgUnitQueryResult(ctx, queryResult, overtimeMetricRange, overtimeMetricType, orgUnitNameLabel, result)
	}
}

func parseOrgUnitQueryResult(ctx context.Context, queryResult metrics_api.Result,
	overtimeMetricRange dao.OvertimeMetricRange, overtimeMetricType OvertimeMetricType,
	orgUnitNameLabel string, result map[string]map[string]*models.OvertimeData,
) {
	if len(queryResult.Measurements) == 0 {
		return
	}

	labels := queryResult.MetricLabels

	clusterId, found := labels[clusterIdLabel]
	if !found {
		return
	}
	orgUnitName, found := labels[orgUnitNameLabel]
	if !found {
		return
	}

	if _, ok := result[clusterId]; !ok {
		result[clusterId] = map[string]*models.OvertimeData{}
	}
	if _, ok := result[clusterId][orgUnitName]; !ok {
		result[clusterId][orgUnitName] = &models.OvertimeData{}
	}

	err := setOvertimeMetricValue(overtimeMetricRange, overtimeMetricType,
		queryResult.Measurements[0].Value, result[clusterId][orgUnitName])
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set overtime metric value for %s %s in cluster %s, error: %s",
			orgUnitNameLabel, orgUnitName, clusterId, err.Error())
	}
}

func setOvertimeMetricValue(overtimeMetricRange dao.OvertimeMetricRange, overtimeMetricType OvertimeMetricType,
	valueStr string, overtimeDataResult *models.OvertimeData,
) error {
	numValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return err
	} else if numValue != numValue {
		// NaN check: if the value is NaN, numValue != numValue will be true,
		// IGNORE THE IDE WARNING.
		// bug: NaN was somehow written to db
		err = fmt.Errorf("NaN value for overtime metric: %s %s", overtimeMetricType, overtimeMetricRange)
		return err
	}
	value := ptr.To(numValue)

	switch overtimeMetricRange {
	case dao.RangeDay:
		switch overtimeMetricType {
		case GpuAllocationAvg:
			overtimeDataResult.AvgGpuAllocation24h = value
		case GpuUtilizationAvg:
			overtimeDataResult.AvgGpuUtilization24h = value
		case GpuMemoryUtilizationAvg:
			overtimeDataResult.AvgGpuMemoryUtilization24h = value
		}
	case dao.RangeWeek:
		switch overtimeMetricType {
		case GpuAllocationAvg:
			overtimeDataResult.AvgGpuAllocation7d = value
		case GpuUtilizationAvg:
			overtimeDataResult.AvgGpuUtilization7d = value
		case GpuMemoryUtilizationAvg:
			overtimeDataResult.AvgGpuMemoryUtilization7d = value
		}
	case dao.RangeMonth:
		switch overtimeMetricType {
		case GpuAllocationAvg:
			overtimeDataResult.AvgGpuAllocation30d = value
		case GpuUtilizationAvg:
			overtimeDataResult.AvgGpuUtilization30d = value
		case GpuMemoryUtilizationAvg:
			overtimeDataResult.AvgGpuMemoryUtilization30d = value
		}
	default:
		return fmt.Errorf("invalid overtime metric range: %s", overtimeMetricRange)
	}

	return nil
}
