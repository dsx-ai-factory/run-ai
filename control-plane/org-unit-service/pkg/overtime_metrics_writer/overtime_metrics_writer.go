/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package overtime_metrics_writer

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"

	org_unit_config "github.com/run-ai/runai/control-plane/org-unit-service/pkg/config"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	tenants_manager_client "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type OvertimeMetricType string

const (
	GpuAllocationAvg        OvertimeMetricType = "gpu_allocation_avg"
	GpuUtilizationAvg       OvertimeMetricType = "gpu_utilization_avg"
	GpuMemoryUtilizationAvg OvertimeMetricType = "gpu_memory_utilization_avg"

	GpuUtilizationPerProjectNodepoolTop15             OvertimeMetricType = "gpu_utilization_per_project_nodepool_top15"
	GpuAllocationPerProjectNodepoolTop15              OvertimeMetricType = "gpu_allocation_per_project_nodepool_top15"
	GpuMemoryUtilizationPerProjectNodepoolTop15       OvertimeMetricType = "gpu_memory_utilization_per_project_nodepool_top15"
	GpuUtilizationPerProjectNodepoolBottom15          OvertimeMetricType = "gpu_utilization_per_project_nodepool_bottom15"
	GpuAllocationPerProjectNodepoolBottom15           OvertimeMetricType = "gpu_allocation_per_project_nodepool_bottom15"
	GpuMemoryUtilizationPerProjectNodepoolBottom15    OvertimeMetricType = "gpu_memory_utilization_per_project_nodepool_bottom15"
	GpuUtilizationPerDepartmentNodepoolTop15          OvertimeMetricType = "gpu_utilization_per_department_nodepool_top15"
	GpuAllocationPerDepartmentNodepoolTop15           OvertimeMetricType = "gpu_allocation_per_department_nodepool_top15"
	GpuMemoryUtilizationPerDepartmentNodepoolTop15    OvertimeMetricType = "gpu_memory_utilization_per_department_nodepool_top15"
	GpuUtilizationPerDepartmentNodepoolBottom15       OvertimeMetricType = "gpu_utilization_per_department_nodepool_bottom15"
	GpuAllocationPerDepartmentNodepoolBottom15        OvertimeMetricType = "gpu_allocation_per_department_nodepool_bottom15"
	GpuMemoryUtilizationPerDepartmentNodepoolBottom15 OvertimeMetricType = "gpu_memory_utilization_per_department_nodepool_bottom15"

	tenantStatusReady = "Ready"
)

var allOvertimeMetricTypes = []OvertimeMetricType{GpuAllocationAvg, GpuUtilizationAvg, GpuMemoryUtilizationAvg}

type OvertimeMetricsWriter struct {
	rateLimitingController *RateLimitingController
	tenantsClient          tenants_manager_client.ITenantsManager
}

func InitGoRoutines(ctx context.Context, stopCh <-chan struct{}, tenantsClient tenants_manager_client.ITenantsManager) {
	go initGoRoutines(ctx, stopCh, tenantsClient)
}

func newOvertimeMetricsWriter(rateLimitingController *RateLimitingController, tenantsClient tenants_manager_client.ITenantsManager) *OvertimeMetricsWriter {
	return &OvertimeMetricsWriter{
		rateLimitingController: rateLimitingController,
		tenantsClient:          tenantsClient,
	}
}

func initGoRoutines(
	ctx context.Context, stopCh <-chan struct{}, tenantsClient tenants_manager_client.ITenantsManager,
) {
	conf := config.Get[*org_unit_config.OvertimeMetrics]()
	rateLimitingController := NewRateLimitingController(conf.RateLimiterLimit, conf.RateLimiterBurst)
	go rateLimitingController.Run(ctx)

	omw := newOvertimeMetricsWriter(rateLimitingController, tenantsClient)

	omw.runOvertimeMetricsWriterPeriodically(ctx, stopCh, dao.RangeDay, conf.Query24hRangeIntervalSeconds, conf.Query24hRangeInitialSleepSeconds)
	omw.runOvertimeMetricsWriterPeriodically(ctx, stopCh, dao.RangeWeek, conf.Query7dRangeIntervalSeconds, conf.Query7dRangeInitialSleepSeconds)
	omw.runOvertimeMetricsWriterPeriodically(ctx, stopCh, dao.RangeMonth, conf.Query30dRangeIntervalSeconds, conf.Query30dRangeInitialSleepSeconds)
}

// runOvertimeMetricsWriterPeriodically runs the overtime metrics writer periodically for the given range and interval.
func (omw *OvertimeMetricsWriter) runOvertimeMetricsWriterPeriodically(ctx context.Context, stopCh <-chan struct{},
	overtimeMetricRange dao.OvertimeMetricRange, intervalSeconds int, initialSleepSeconds int,
) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Setting up a go routine of overtime metrics writer for range %v, interval: %d s, initial sleep: %d s",
		overtimeMetricRange, intervalSeconds, initialSleepSeconds)

	interval := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)

	time.Sleep(time.Duration(initialSleepSeconds) * time.Second)

	go func() {
		omw.runOvertimeMetricsWriter(ctx, overtimeMetricRange)
		for {
			select {
			case <-ticker.C:
				omw.runOvertimeMetricsWriter(ctx, overtimeMetricRange)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// runOvertimeMetricsWriter - runs the overtime metrics writer for the given range.
// for each tenant with a non-deleted org unit, adds an item to the rate limiting controller.
func (omw *OvertimeMetricsWriter) runOvertimeMetricsWriter(ctx context.Context, overtimeMetricRange dao.OvertimeMetricRange) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Running overtime metrics writer for range %v", overtimeMetricRange)

	tenants, err := omw.getTenants(ctx)
	if err != nil {
		return
	}
	sort.Strings(tenants)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Found %d tenants with non-deleted org units: %v",
		len(tenants), tenants)

	for _, tenantId := range tenants {
		omw.rateLimitingController.AddRateLimited(RateLimitedItem{
			tenantId:            tenantId,
			overtimeMetricRange: overtimeMetricRange,
		})
	}
}

func (omw *OvertimeMetricsWriter) getTenants(ctx context.Context) ([]string, error) {
	tenants, resp, err := omw.tenantsClient.ListAllTenantsInternal(context.Background())
	if err != nil || resp == nil {
		return nil, fmt.Errorf("failed to list tenants: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list tenants. resp: %v, status code: %d", resp.Body, resp.StatusCode)
	}

	result := []string{}
	for _, tenant := range tenants {
		if tenant.Status == tenantStatusReady {
			result = append(result, strconv.Itoa(int(tenant.Id)))
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("skipping addition of tenant %s (id %d) - tenant status not ready: %s",
				tenant.Name, tenant.Id, tenant.Status)
		}
	}

	return result, nil
}

// QueryOvertimeMetrics -
// queries the overtime metrics for the given tenant and range for both projects and departments,
// and writes the results to the database.
func QueryOvertimeMetrics(ctx context.Context, tenantId string, overtimeMetricRange dao.OvertimeMetricRange) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Querying overtime metrics for tenant %s, range %v", tenantId, overtimeMetricRange)

	result, err := queryOvertimeMetricsForOrgUnit(ctx, tenantId, overtimeMetricRange, true)
	if err != nil {
		return err
	}

	err = dao.SetProjectsOvertimeDataForTenant(ctx, tenantId, result, overtimeMetricRange)
	if err != nil {
		return err
	}

	result, err = queryOvertimeMetricsForOrgUnit(ctx, tenantId, overtimeMetricRange, false)
	if err != nil {
		return err
	}

	return dao.SetDepartmentsOvertimeDataForTenant(ctx, tenantId, result, overtimeMetricRange)
}

// queryOvertimeMetricsForOrgUnit -
// queries all the overtime metrics for the given tenant and range for either projects or departments.
// returns a map of clusterId -> orgUnitName -> overtimeData.
func queryOvertimeMetricsForOrgUnit(ctx context.Context, tenantId string, overtimeMetricRange dao.OvertimeMetricRange, queryProjects bool) (
	map[string]map[string]*models.OvertimeData, error,
) {
	result := map[string]map[string]*models.OvertimeData{}

	for _, overtimeMetricType := range allOvertimeMetricTypes {
		err := queryOvertimeMetric(ctx, tenantId, overtimeMetricRange, overtimeMetricType, queryProjects, result)
		if err != nil {
			return nil, err
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully queried all metrics for tenant: %s, range: %v, queryProjects: %t",
		tenantId, overtimeMetricRange, queryProjects)
	return result, nil
}

// queryOvertimeMetric -
// queries the specific overtime metric for the given tenant and range for either projects or departments.
// saves the results to the given result map.
func queryOvertimeMetric(ctx context.Context, tenantId string, overtimeMetricRange dao.OvertimeMetricRange,
	overtimeMetricType OvertimeMetricType, queryProjects bool,
	result map[string]map[string]*models.OvertimeData,
) error {
	queryResults, err := getOvertimeMetricResults(ctx, tenantId, overtimeMetricRange, overtimeMetricType, queryProjects)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to get metric results for tenant %s, overtimeMetricRange: %v, overtimeMetricType: %v, queryProjects: %t, error: %v",
			tenantId, overtimeMetricRange, overtimeMetricType, queryProjects, err)
		return err
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully queried metric for tenant %s, overtimeMetricRange: %v, overtimeMetricType: %v, queryProjects: %t",
		tenantId, overtimeMetricRange, overtimeMetricType, queryProjects)

	parseTenantOvertimeData(ctx, queryResults, overtimeMetricRange, overtimeMetricType, queryProjects, result)
	return nil
}

func QueryProjectsOvertimeRankingMetrics(ctx context.Context, tenantId string, metricType api.RankingMetricType, orderType api.RankingOrderType, clusterFilter []string, nodePoolName string, departmentFilter []string) (
	[]api.ProjectGetRanking, error,
) {
	var nodePoolFilter []string
	if nodePoolName != "" {
		nodePoolFilter = []string{nodePoolName}
	}
	metricResults, overtimeMetricRange, err := queryOvertimeRankingMetrics(ctx, tenantId, metricType, orderType, clusterFilter, nodePoolFilter, departmentFilter, true)
	if err != nil {
		return nil, err
	}

	return mapOvertimeDataToProjectGetRanking(ctx, tenantId, metricResults, overtimeMetricRange, nodePoolName)
}

func QueryDepartmentsOvertimeRankingMetrics(ctx context.Context, tenantId string, metricType api.RankingMetricType,
	orderType api.RankingOrderType, clusterFilter []string, nodePoolName string, departmentFilter []string,
) (
	[]api.DepartmentGetRanking, error,
) {
	var nodePoolFilter []string
	if nodePoolName != "" {
		nodePoolFilter = []string{nodePoolName}
	}
	results, overtimeMetricRange, err := queryOvertimeRankingMetrics(ctx, tenantId, metricType, orderType, clusterFilter, nodePoolFilter, departmentFilter, false)
	if err != nil {
		return nil, err
	}

	return mapOvertimeDataToDepartmentGetRanking(ctx, tenantId, results, overtimeMetricRange, nodePoolName)
}

func queryOvertimeRankingMetrics(ctx context.Context, tenantId string, metricType api.RankingMetricType, orderType api.RankingOrderType, clusterFilter []string,
	nodePoolFilter []string, departmentFilter []string, queryProjects bool) (
	map[string]map[string]*models.OvertimeData, dao.OvertimeMetricRange, error,
) {
	var secondaryDepartmentFilter []string
	var secondaryProjectFilter []string
	results := map[string]map[string]*models.OvertimeData{}

	overtimeMetricRange, err := getTimeRangeForRankingMetric(metricType)
	if err != nil {
		return nil, "", err
	}

	metricsList, err := getRankingMetricsList(metricType, orderType, queryProjects)
	if err != nil {
		return nil, "", err
	}

	err = queryOvertimeMetricDataForRanking(ctx, tenantId, overtimeMetricRange, metricsList[0], clusterFilter, nodePoolFilter, departmentFilter, []string{}, queryProjects, results)
	if err != nil {
		return nil, "", err
	}

	if queryProjects {
		secondaryProjectFilter = extractOrgUnitNamesFromOverTimeData(results)
	} else {
		secondaryDepartmentFilter = extractOrgUnitNamesFromOverTimeData(results)
	}

	for i := 1; i < len(metricsList); i++ {
		err = queryOvertimeMetricDataForRanking(ctx, tenantId, overtimeMetricRange, metricsList[i], clusterFilter, nodePoolFilter, secondaryDepartmentFilter, secondaryProjectFilter, queryProjects, results)
		if err != nil {
			return nil, "", err
		}
	}

	return results, overtimeMetricRange, nil
}
