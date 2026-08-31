/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authorization "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

func DbUpsertWorkloadV2Snapshot(ctx context.Context, snapshot *models_crud.WorkloadV2Snapshot) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to upsert V2 workload snapshot %s", snapshot.ID), nil)

	result := dao_utils.DB().WithContext(ctx).Table(models_crud.WorkloadV2SnapshotsTableName).Save(snapshot)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert V2 workload snapshot %s, err %v", snapshot.ID, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbGetWorkloadV2Snapshot(ctx context.Context, tenantId uint, workloadType string, workloadId string,
	permittedScopes *authorization.PermittedScopes,
) (*models_crud.WorkloadV2Snapshot, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get V2 workload snapshot %s %s", workloadType, workloadId), nil)
	notFoundErr := api.NewError(http.StatusNotFound, fmt.Sprintf("V2 workload snapshot %s %s not found", workloadType, workloadId), nil)

	var snapshot models_crud.WorkloadV2Snapshot

	result := dao_utils.DB().WithContext(ctx).
		Model(snapshot).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnId), workloadId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnType), workloadType)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	result = result.First(&snapshot)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, notFoundErr.WithDetails(result.Error)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get V2 workload snapshot %s with id %s, err: %v", workloadType, workloadId, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &snapshot, nil
}

func DbDeleteWorkloadV2Snapshot(ctx context.Context, tenantId uint, workloadId string, permittedScopes *authorization.PermittedScopes) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete V2 workload snapshot %s", workloadId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadV2SnapshotsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnId), workloadId)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	result = result.Delete(&models_crud.WorkloadV2Snapshot{})
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete V2 workload snapshot %s, err %v", workloadId, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbDeleteWorkloadV2Snapshots(ctx context.Context, tenantId uint, workloadIds []string, permittedScopes *authorization.PermittedScopes) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete V2 workload snapshots", nil)

	if len(workloadIds) == 0 {
		return nil
	}

	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadV2SnapshotsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s IN ?", models_crud.WorkloadV2SnapshotsColumnId), workloadIds)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	result = result.Delete(&models_crud.WorkloadV2Snapshot{})
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete V2 workload snapshots, err %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}
