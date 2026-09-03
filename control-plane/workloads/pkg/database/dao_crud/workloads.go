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
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/google/uuid"

	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func DbUpsertWorkload(ctx context.Context, workload *models_crud.Workload) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to upsert workload %s", workload.Name), nil)

	result := dao_utils.DB().WithContext(ctx).Table(models_crud.WorkloadsTableName).Save(&workload)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert workload %s, err %v", workload.DebugInfo(), result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

func DbUpdateDesiredPhase(ctx context.Context, tenantId uint, workloadType string, workloadId string,
	permittedScopes *authorization.PermittedScopes, desiredPhase api.WorkloadDesiredPhase,
) (*time.Time, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update desired phase of %s %s", workloadType, workloadId), nil)

	updateTime := time.Now()
	updatedBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	result := dao_utils.DB().WithContext(ctx).Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnId), workloadId)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	result = result.Update(models_crud.WorkloadsColumnDesiredPhase, desiredPhase).
		Update(models_crud.WorkloadsColumnUpdatedAt, updateTime).
		Update(models_crud.WorkloadsColumnUpdatedBy, updatedBy)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update desired state of %s %s, err %v", workloadType, workloadId, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, api.NewError(http.StatusNotFound, "workload not found", nil)
	}

	return &updateTime, nil
}

func DbAddChangeStateRecord(ctx context.Context, workloadType string, workload *models.WorkloadView, desiredPhase api.WorkloadDesiredPhase) (*time.Time, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to add state-change %s for %s", workloadType, workload.ID), nil)

	updatedBy := runaictx.GetString(ctx, runaictx.CfSubjectID)
	updateTime := time.Now()

	imageSpec := workloadsApi.NullableString{}
	if len(workload.Images) > 0 {
		images := []string(workload.Images)
		imageSpec = *workloadsApi.NewNullableString(&images[0])
	}

	crudWorkload := models_crud.Workload{
		ID:           workload.ID,
		CreatedAt:    updateTime,
		UpdatedAt:    updateTime,
		TenantId:     int32(workload.TenantID),
		ClusterId:    workload.ClusterID.String(),
		DepartmentId: workload.DepartmentID,
		ProjectId:    workload.ProjectID,
		WorkloadType: workloadType,
		Name:         workload.Name,
		Namespace:    workload.Namespace,
		DesiredPhase: convutils.Ref(workloadsApi.WorkloadDesiredPhase(desiredPhase)),
		Spec: &workloadsApi.SupersetSpec{
			Image: imageSpec,
		},
	}

	if updatedBy != nil {
		crudWorkload.UpdatedBy = *updatedBy
		crudWorkload.CreatedBy = *updatedBy
	}

	result := dao_utils.DB().WithContext(ctx).Table(models_crud.WorkloadsTableName).Save(&crudWorkload)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to add state-change record %s, err %v", crudWorkload.DebugInfo(), result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return &updateTime, nil
}

func DbGetWorkload(ctx context.Context, tenantId uint, workloadType string, workloadId string,
	permittedScopes *authorization.PermittedScopes,
) (*models_crud.Workload, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to get %s %s", workloadType, workloadId), nil)

	var workload models_crud.Workload

	result := dao_utils.DB().WithContext(ctx).
		Model(workload).
		Unscoped().
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnId), workloadId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnType), workloadType)

	addPermittedScopesQueries(result, tenantId, permittedScopes)

	result = result.First(&workload)

	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %s with id %s, err: %v", workloadType, workloadId, result.Error)
		}
		return nil, dbErr.WithDetails(result.Error)
	}

	return &workload, nil
}

func DbListUpdates(ctx context.Context, tenantId uint, workloadType string, clusterId string, afterUpdate int64, bulkSize int) ([]models_crud.WorkloadSyncUpdate, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to list %s updates for cluster %s", workloadType, clusterId), nil)

	var records []models_crud.WorkloadSyncUpdate

	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnType), workloadType).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnClusterId), clusterId).
		Where(fmt.Sprintf("%s > ?", models_crud.WorkloadsColumnUpdateCount), afterUpdate).
		Limit(bulkSize).
		Order(models_crud.WorkloadsColumnUpdateCount).
		Scan(&records)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list %s updates for cluster %s, err: %v", workloadType, clusterId, result.Error)
		return nil, dbErr.WithDetails(result.Error)
	}

	return records, nil
}

func addPermittedScopesQueries(tx *gorm.DB, tenantId uint, permittedScopes *authorization.PermittedScopes) {
	if permittedScopes == nil {
		return
	}

	if (permittedScopes.System == nil || !*permittedScopes.System) && !slices.Contains(permittedScopes.Tenants, fmt.Sprintf("%d", tenantId)) {
		orExpression := "(false"

		if len(permittedScopes.Clusters) > 0 {
			orExpression += " or cluster_id IN " + dao_utils.ListItemsForQuery(permittedScopes.Clusters)
		}

		if len(permittedScopes.Departments) > 0 {
			orExpression += " or department_id IN " + dao_utils.ListItemsForQuery(permittedScopes.Departments)
		}

		if len(permittedScopes.Projects) > 0 {
			orExpression += " or project_id IN " + dao_utils.ListItemsForQuery(permittedScopes.Projects)
		}

		orExpression += ")"

		tx.Where(orExpression)
	}
}

func CleanupObsoleteWorkloadUpdates(ctx context.Context, tenantId uint, workloadType string, clusterId string, afterUpdate int64) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete obsolete %s updates for cluster %s", workloadType, clusterId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnClusterId), clusterId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnType), workloadType).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnDesiredPhase), api.WORKLOADDESIREDPHASE_DELETED).
		Where(fmt.Sprintf("%s < ?", models_crud.WorkloadsColumnUpdateCount), afterUpdate).
		Delete(&models_crud.WorkloadSyncUpdate{})

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete obsolete %s updates for cluster %s, err: %v", workloadType, clusterId, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}

// DeleteWorkloadsOfProjectId deletes workloads of a project
func DeleteWorkloadsOfProjectId(ctx context.Context, tenantId string, clusterId string, projectId string) error {
	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnClusterId), clusterId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnProjectId), projectId).
		Update(models_crud.WorkloadsColumnDesiredPhase, api.WORKLOADDESIREDPHASE_DELETED)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workload for project %s in tenant %s. err: %v", projectId, tenantId, result.Error)
		return fmt.Errorf("failed to delete workload for project %s in tenant %s. err: %w", projectId, tenantId, result.Error)
	}

	return nil
}

// DeleteWorkloadsOfTenantId cleanup workloads of a tenant from the wl-crud
func DeleteWorkloadsOfTenantId(ctx context.Context, tenantId string) error {
	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Delete(&models_crud.WorkloadSyncUpdate{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete workload in tenant %s. err: %w", tenantId, result.Error)
	}
	return nil
}

// DeleteWorkloadsOfClusterId cleanup workloads of a cluster from the wl-crud
func DeleteWorkloadsOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	result := dao_utils.DB().WithContext(ctx).
		Table(models_crud.WorkloadsTableName).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s = ?", models_crud.WorkloadsColumnClusterId), clusterId).
		Delete(&models_crud.WorkloadSyncUpdate{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete workload for cluster %s in tenant %s. err: %w", clusterId, tenantId, result.Error)
	}

	return nil
}

// ============================================================================================
// The functions below are intended for receiving callbacks from workloads -> to -> workloads-crud.
// The source of truth for this data is the data of the workloads (as opposed to workloads_crud).

// DeleteWorkloadById is invoked when the workloads part receives information from the cluster sync that a workload has been deleted in the cluster.
// The reason it is relevant to us, is this case. (1) a workload that was created by the workloads-crud -> agent path (2) the workload has been removed in the cluster itself, bypassing the workloads-crud -> agent path.
// In such case we want to remove it from workloads-crud, to prevent it from being re-created by the agent in case it issue a get update with update-count smaller than the update count of the workload's creation.
func DeleteWorkloadById(ctx context.Context, id uuid.UUID) error {
	tenantId := auth.TenantIdOf(ctx)
	workloadCrudRec := &models_crud.Workload{
		ID:       id,
		TenantId: tenantId,
	}
	result := dao_utils.DB().WithContext(ctx).Unscoped().Model(&workloadCrudRec).Where(workloadCrudRec).Delete(dao_utils.All)
	if result.Error != nil {
		return fmt.Errorf("failed to delete workload with id %s in tenant %d. err: %w", id, tenantId, result.Error)
	} else if result.RowsAffected > 0 {
		logging.Infof("DeleteWorkloadById callback removed workload %s of tenant %d, following an update that is has been removed from the cluster", id.String(), tenantId)
	}

	return nil
}

// DbSyncWorkloads - callback is invoked by the workloads mechanism for syncing with the cluster. our goal is to remove from our table records that are no longer in the cluster.
// to/from is the range of workload ids that were sent from the caller, therefore we can restrict the search to this range.
// it is premised that the caller is working in ascending uuid order, therefore the range of from/to can restrict the query accordingly.
// existingWorkloadIds is an array of the ids of workloads that were found in the cluster, and therefore we should not delete.
// deletionFilters is filters that we get only from clusters older than 2.16, when deployments and the rest of the workloads didn't sync together so we don't want to mistakenly delete the other type of the workloads
func DbSyncWorkloads(
	ctx context.Context,
	tenantId uint,
	projectId string,
	clusterId string,
	to *uuid.UUID,
	from *uuid.UUID,
	existingWorkloadIds []string,
	deletionFilters []validators.ParameterizedFilter,
) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync workloads", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	if projectId == "" && clusterId == "" {
		return api.NewError(http.StatusBadRequest, "workloads sync requires either projectId or clusterId", nil)
	}

	deleteQuery := dao.ConstructDeleteQueryForProjectCluster(tx, models_crud.WorkloadsTableName, tenantId, clusterId, projectId, from, to, existingWorkloadIds, "")

	if len(deletionFilters) > 0 {
		deleteQuery = dao_utils.ApplyParameterizedFilters(deleteQuery, deletionFilters)
	}

	logging.Infof("applying the following deletion query based on sync with the cluster:")
	result := deleteQuery.Unscoped().
		Delete(&models_crud.Workload{})

	if result.Error != nil {
		return dbErr.WithDetails(result.Error)
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the workloads sync: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}
