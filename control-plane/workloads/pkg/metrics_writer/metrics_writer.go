/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_writer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"

	errors_utils "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/utils/errors"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/runaimetrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

const (
	WorkloadsTableName = "workloads"
	RunningPhase       = "Running"
	workloadIdLabel    = "workload_id"
)

type WorkloadData struct {
	Timestamp         time.Time
	IdleAllocatedGpus float32
}

func InitGoRoutines(ctx context.Context, stopCh <-chan struct{}) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Init go routine for telemetry data update")

	conf := runai_config.Get[*config.MetricsWriter]()
	intervalMinutes := conf.IntervalMinutes
	if intervalMinutes <= 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get metrics-writer config")
		return
	}
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)

	go func() {
		updateRunningWorkloadPerTenant(ctx, conf, intervalMinutes*60)
		for {
			select {
			case <-ticker.C:
				updateRunningWorkloadPerTenant(ctx, conf, intervalMinutes*60)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func updateRunningWorkloadPerTenant(ctx context.Context, conf *config.MetricsWriter, intervalSeconds int) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("update running workload per tenant")
	db := dao_utils.DB()
	tenantIds, err := getTenantsWithRunningWorkloads(ctx, db)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenants, err %v", err)
		return
	}
	tenantsNum := len(tenantIds)
	if tenantsNum == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("there are no tenants with running workloads")
		return
	}
	tenantInterval, tenantsChunk := getTenantsIntervalAndChunk(conf.MinimumTenantIntervalSeconds, tenantsNum, intervalSeconds)
	if tenantsChunk > conf.MaximumTenantsInChunk {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("there are too many tenants for metrics call. number of tenants: %v, interval: %v minutes", tenantsNum, conf.IntervalMinutes)
		return
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("update workload per tenant in chunks of %v every %v seconds", tenantsChunk, tenantInterval)

	ticker := time.NewTicker(time.Duration(tenantInterval) * time.Second)
	defer ticker.Stop()

	for i := 0; i < tenantsNum; i += tenantsChunk {
		end := i + tenantsChunk

		if end > tenantsNum {
			end = tenantsNum
		}
		tenantIdsChunk := tenantIds[i:end]
		<-ticker.C
		updateWorkloadIdleGpus(ctx, db, tenantIdsChunk)
	}
}

func getTenantsIntervalAndChunk(minimumTenantIntervalSeconds int, tenantsNum int, intervalSeconds int) (int, int) {
	tenantInterval := intervalSeconds / tenantsNum
	tenantsChunk := 1
	if tenantInterval < minimumTenantIntervalSeconds {
		// in case there are too many tenants, run in chunks (edge case)
		tenantInterval = minimumTenantIntervalSeconds
		tenantsChunk = (tenantsNum * minimumTenantIntervalSeconds / intervalSeconds) + 1
	}
	return tenantInterval, tenantsChunk
}

func updateWorkloadIdleGpus(ctx context.Context, db *gorm.DB, tenantIds []string) {
	logging.Debugf("updating non running workloads to be with null idle gpu number")
	err := updateNotRunningWorkloads(ctx, db, tenantIds)
	if err != nil {
		logging.Errorf("failed to update not running workloads idle gpu number, err %v", err)
		return
	}
	runningWorkloadIds, err := getRunningWorkloads(ctx, db, tenantIds)
	if err != nil {
		if !errors_utils.IsNotFound(err) {
			logging.Errorf("failed to get running workloads for idle gpu number update, err %v", err)
		}
		return
	}
	workloadsData, err := getIdleGpusFromMetrics(ctx, runningWorkloadIds)
	if err != nil {
		if !errors_utils.IsNotFound(err) {
			logging.Errorf("failed to get running workloads idle gpu number metrics, err %v", err)
		}
		return
	}
	logging.Debugf("updating running workloads idle gpu number")
	failedWorkloads := updateRunningWorkloads(db, workloadsData, runningWorkloadIds)
	if len(failedWorkloads) > 0 {
		logging.Errorf("failed to update running workloads idle gpu number, failed ids: %v", failedWorkloads)
	}
}

func getTenantsWithRunningWorkloads(ctx context.Context, db *gorm.DB) ([]string, error) {
	tx := db.WithContext(ctx)

	tenantIds := []string{}

	result := tx.Table(WorkloadsTableName).
		Where("phase = ?", RunningPhase).Distinct("tenant_id").Pluck("tenant_id", &tenantIds)

	return tenantIds, result.Error
}

func updateNotRunningWorkloads(ctx context.Context, db *gorm.DB, tenantIds []string) error {
	tx := db.WithContext(ctx)

	result := tx.Table(WorkloadsTableName).
		Where("deleted_at is null").
		Where("tenant_id in ?", tenantIds).
		Where("not phase = ?", RunningPhase).
		Where("(idle_since is not null OR idle_allocated_gpus is not null)").
		Updates(map[string]interface{}{
			"idle_allocated_gpus": nil,
			"idle_since":          nil,
		})
	return result.Error
}

func updateRunningWorkloads(db *gorm.DB, workloadsData map[string]WorkloadData, runningWorkloadIds map[string][]string) []string {
	failedWorkloads := []string{}
	for tenantId, runningWorkloadsPerTenant := range runningWorkloadIds {
		for _, runningWorkload := range runningWorkloadsPerTenant {
			workloadData, found := workloadsData[runningWorkload]
			if !found {
				workloadData = WorkloadData{
					Timestamp:         time.Now(),
					IdleAllocatedGpus: 0,
				}
			}

			updates := map[string]interface{}{
				"idle_allocated_gpus": workloadData.IdleAllocatedGpus,
			}

			if workloadData.IdleAllocatedGpus > 0 {
				updates["idle_since"] = gorm.Expr("COALESCE(idle_since, ?)", workloadData.Timestamp)
			} else {
				updates["idle_since"] = nil
			}

			result := db.Table(WorkloadsTableName).Where("id = ?", runningWorkload).
				Where("tenant_id = ?", tenantId).
				Updates(updates)
			if result.Error != nil {
				failedWorkloads = append(failedWorkloads, runningWorkload)
			}
		}
	}

	return failedWorkloads
}

func getRunningWorkloads(ctx context.Context, db *gorm.DB, tenantIds []string) (map[string][]string, error) {
	tx := db.WithContext(ctx)

	workloadIds := make(map[string][]string)
	var workloads []struct {
		Id       string
		TenantId string
	}
	result := tx.Table(WorkloadsTableName).
		Where("phase = ?", RunningPhase).
		Where("tenant_id in ?", tenantIds).
		Select("id, tenant_id").
		Find(&workloads)

	for _, workload := range workloads {
		ids, found := workloadIds[workload.TenantId]
		if !found {
			ids = []string{}
		}
		ids = append(ids, workload.Id)
		workloadIds[workload.TenantId] = ids
	}
	return workloadIds, result.Error
}

func getIdleGpusFromMetrics(ctx context.Context, workloads map[string][]string) (map[string]WorkloadData, error) {
	workloadsData := make(map[string]WorkloadData)
	for tenantId, workloadIds := range workloads {
		result, err := getIdleGpusMetricsResult(ctx, workloadIds, tenantId)
		if err != nil {
			return nil, err
		}
		tenantWorkloadsData, err := mapMetricsToWorkloads(ctx, result.GetData())
		if err != nil {
			return nil, err
		}
		for workloadId, workloadData := range tenantWorkloadsData {
			workloadsData[workloadId] = workloadData
		}
	}

	return workloadsData, nil
}

func mapMetricsToWorkloads(ctx context.Context, metricsResults []openapi.Result) (map[string]WorkloadData, error) {
	workloadsData := make(map[string]WorkloadData)
	for _, res := range metricsResults {
		workloadId := res.MetricLabels[workloadIdLabel]
		if workloadId == "" {
			return workloadsData, errors.New("cannot get workload id from metric result")
		}
		if len(res.Measurements) != 1 {
			return workloadsData, fmt.Errorf("cannot get metric result for workload %s", workloadId)
		}
		value := res.Measurements[0].Value
		gpuNum, err := strconv.ParseFloat(value, 32)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse idle gpu metric value %v. err: %+v", gpuNum, err)
			return workloadsData, err
		}
		workloadsData[workloadId] = WorkloadData{
			Timestamp:         res.Measurements[0].Timestamp,
			IdleAllocatedGpus: float32(gpuNum),
		}
	}
	return workloadsData, nil
}

func getIdleGpusMetricsResult(ctx context.Context, workloadIds []string, tenantId string) (*openapi.GetMetrics200Response, error) {
	metric := runaimetrics.RunaiIdleGpuCountPerWorkloadMetric{Filters: runaimetrics.RunaiIdleGpuCountPerWorkloadFilters{
		WorkloadId: workloadIds,
	}}

	tenantUid, err := strconv.ParseInt(tenantId, 10, 32)
	if err != nil {
		return nil, err
	}

	query, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		return nil, err
	}

	query.SetOverrideTenantId(int32(tenantUid))

	res, err := query.Query(ctx)
	return res, err
}
