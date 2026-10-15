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

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

func DbUpsertWorkloadV2Snapshot(ctx context.Context, snapshot *models_crud.WorkloadV2Snapshot) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to upsert V2 workload snapshot %s", snapshot.ID), nil)

	stored := *snapshot
	stored.Spec = gpucompat.LegacyWorkloadSpec(snapshot.Spec)
	stored.MasterSpec = gpucompat.LegacyMasterSpec(snapshot.MasterSpec)

	_, err := runaidb.DB().WithContext(ctx).Table(models_crud.WorkloadV2SnapshotsTableName).Save(&stored)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert V2 workload snapshot %s, err %v", snapshot.ID, err)
		return dbErr.WithDetails(err)
	}

	stored.Spec, stored.MasterSpec = snapshot.Spec, snapshot.MasterSpec
	*snapshot = stored

	return nil
}

func DbGetWorkloadV2Snapshot(ctx context.Context, tenantId uint, workloadType string, workloadId string,
	permittedScopes *authorization.PermittedScopes,
) (*models_crud.WorkloadV2Snapshot, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get V2 workload snapshot %s %s", workloadType, workloadId), nil)
	notFoundErr := api.NewError(http.StatusNotFound, fmt.Sprintf("V2 workload snapshot %s %s not found", workloadType, workloadId), nil)

	var snapshot models_crud.WorkloadV2Snapshot

	result := runaidb.DB().WithContext(ctx).
		Model(snapshot).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnId), workloadId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnType), workloadType)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	err := result.First(&snapshot)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFoundErr.WithDetails(err)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get V2 workload snapshot %s with id %s, err: %v", workloadType, workloadId, err)
		return nil, dbErr.WithDetails(err)
	}

	return &snapshot, nil
}

func DbDeleteWorkloadV2Snapshot(ctx context.Context, tenantId uint, workloadId string, permittedScopes *authorization.PermittedScopes) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete V2 workload snapshot %s", workloadId), nil)

	result := runaidb.DB().WithContext(ctx).
		Table(models_crud.WorkloadV2SnapshotsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnId), workloadId)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	_, err := result.Delete(&models_crud.WorkloadV2Snapshot{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete V2 workload snapshot %s, err %v", workloadId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DbDeleteWorkloadV2Snapshots(ctx context.Context, tenantId uint, workloadIds []string, permittedScopes *authorization.PermittedScopes) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete V2 workload snapshots", nil)

	if len(workloadIds) == 0 {
		return nil
	}

	result := runaidb.DB().WithContext(ctx).
		Table(models_crud.WorkloadV2SnapshotsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadV2SnapshotsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s IN ?", models_crud.WorkloadV2SnapshotsColumnId), workloadIds)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	_, err := result.Delete(&models_crud.WorkloadV2Snapshot{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete V2 workload snapshots, err %v", err)
		return dbErr.WithDetails(err)
	}

	return nil
}
