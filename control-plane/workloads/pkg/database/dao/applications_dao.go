/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const UniqueApplicationConstraint = "applications_unique"

func SyncClusterApplications(ctx context.Context, tenantId string, clusterId uuid.UUID, applicationsForUpsert []models.Application, applicationIdsForDelete []string, syncDeleted bool) ([]string, []string, error) {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return nil, nil, tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	synced := []string{}
	failed := []string{}
	excludedIdsFromDelete := []string{}

	if len(applicationsForUpsert) > 0 {
		syncedExisting, failedExisting, ids := upsertApplications(tx, applicationsForUpsert)
		synced = append(synced, syncedExisting...)
		failed = append(failed, failedExisting...)
		excludedIdsFromDelete = append(excludedIdsFromDelete, ids...)
	}
	if len(applicationIdsForDelete) > 0 {
		err := deleteApplications(tx, tenantId, clusterId.String(), applicationIdsForDelete)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete applications. err %v", err)
			failed = append(failed, applicationIdsForDelete...)
		} else {
			tx.SavePoint("sp")
			synced = append(synced, applicationIdsForDelete...)
		}
	}

	if syncDeleted {
		err := deleteApplicationsWithExcluded(tx, tenantId, clusterId.String(), excludedIdsFromDelete)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete priorities full sync. err %v", err)
		} else {
			tx.SavePoint("sp")
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit priorities sync: %v", result.Error)
		return nil, nil, result.Error
	}
	committed = true

	return synced, failed, nil
}

func deleteApplications(tx *gorm.DB, tenantId string, clusterId string, applicationIdsForDelete []string) error {
	result := tx.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("id in ?", applicationIdsForDelete).
		Delete(&models.Application{})

	return result.Error
}

func deleteApplicationsWithExcluded(tx *gorm.DB, tenantId string, clusterId string, applicationIdsToExclude []string) error {
	// in case of full sync, we need to delete all applications which are not received from cluster
	query := tx.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId)

	if len(applicationIdsToExclude) > 0 {
		query = query.Where("id not in ?", applicationIdsToExclude)
	}

	result := query.Delete(&models.Application{})

	return result.Error
}

func upsertApplications(tx *gorm.DB, applications []models.Application) ([]string, []string, []string) {
	synced := []string{}
	failed := []string{}
	ids := []string{}
	for _, application := range applications {
		// Omit CreatedAt/CreatedBy: the sync model does not carry them, and Save updates
		// all columns on an existing row - without this it would overwrite created_at/
		// created_by with zero values on every re-sync. On insert these fall back to the
		// column defaults (created_at DEFAULT now()).
		saveResult := tx.Table(models.ApplicationTableName).Omit("CreatedAt", "CreatedBy").Save(&application)
		if saveResult.Error != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			pqErr := getPgError(saveResult.Error)
			if pqErr != nil && pqErr.Code == UniqueViolationErr && pqErr.ConstraintName == UniqueApplicationConstraint {
				// application already exists - update
				resources, err := json.Marshal(application.Resources)
				if err != nil {
					failed = append(failed, application.ID.String())
					continue
				}
				saveResultErr := tx.Table(models.ApplicationTableName).
					Where("id = ?", application.ID).
					Where("tenant_id = ?", application.TenantId).
					Where("cluster_id = ?", application.ClusterId).
					Updates(map[string]interface{}{
						"updated_at": time.Now(),
						"k8s_phase":  application.K8sPhase,
						"resources":  resources,
					})
				if saveResultErr != nil {
					tx.RollbackTo("sp")
					failed = append(failed, application.ID.String())
				} else {
					tx.SavePoint("sp")
					synced = append(synced, application.ID.String())
				}
			} else {
				failed = append(failed, application.ID.String())
			}
		} else {
			tx.SavePoint("sp") // save point
			synced = append(synced, application.ID.String())
		}
		ids = append(ids, application.ID.String())
	}
	return synced, failed, ids
}

func applyQueryParams(query *gorm.DB, offset int32, limit int32, sortOrder string, sortBy string, filterBy []validators.ParameterizedFilter) (*gorm.DB, error) {
	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	var err error
	if sortBy != "" {
		query, err = dao_utils.ApplySortByParam(query, sortBy, sortOrder)
		if err != nil {
			return query, err
		}
	}
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	return query, nil
}

func GetApplications(
	ctx context.Context,
	tenantId string,
	permittedScopes *openapi.PermittedScopes,
	offset int32,
	limit int32,
	sortOrder string,
	sortBy string,
	filterBy []validators.ParameterizedFilter,
	search string,
	verboseView bool,
) (*[]models.ApplicationView, error) {
	query := dao_utils.DB().WithContext(ctx)
	if verboseView {
		query = query.Table(models.ApplicationViewName)
	} else {
		query = query.
			Table(models.ApplicationTableName).
			Select(fmt.Sprintf("%s.*, acs.spec, acs.desired_phase, acs.%s", models.ApplicationTableName, models.ApplicationCrudSyncColumnUpdateCount)).
			Joins(fmt.Sprintf("LEFT JOIN (SELECT %s, spec, %s, %s FROM %s) AS acs ON acs.%s = %s.%s",
				models.ApplicationCrudSyncColumnAppId,
				models.ApplicationColumnDesiredPhase,
				models.ApplicationCrudSyncColumnUpdateCount,
				models.ApplicationCrudSyncTableName,
				models.ApplicationCrudSyncColumnAppId,
				models.ApplicationTableName,
				models.ApplicationColumnId))
	}
	query = query.Where(fmt.Sprintf("%s = ?", models.ApplicationColumnTenantId), tenantId)

	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		// if user is not system - filter by permitted scopes
		scopesCond, scopes := getPermissionQuery(permittedScopes)
		query = query.Where(scopesCond, scopes...)
	}

	var err error
	query, err = applyQueryParams(query, offset, limit, sortOrder, sortBy, filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to apply query params to applications, err: %v", err)
		return &[]models.ApplicationView{}, err
	}

	if search != "" {
		query = dao_utils.ApplyTSSearchParam(query, search, "search_vector")
	}

	applications := []models.ApplicationView{}
	result := query.Find(&applications)
	if result.Error != nil {
		logging.Errorf("Failed to get applications: %v", result.Error)
		return nil, result.Error
	}
	return &applications, nil
}

func getPermissionQuery(permittedScopes *openapi.PermittedScopes) (string, []interface{}) {
	query := "false"
	values := []interface{}{}
	if len(permittedScopes.Tenants) > 0 {
		query = fmt.Sprintf("%s or tenant_id in ?", query)
		values = append(values, permittedScopes.Tenants)
	}
	if len(permittedScopes.Clusters) > 0 {
		query = fmt.Sprintf("%s or cluster_id in ?", query)
		values = append(values, permittedScopes.Clusters)
	}
	if len(permittedScopes.Departments) > 0 {
		query = fmt.Sprintf("%s or department_id in ?", query)
		values = append(values, permittedScopes.Departments)
	}
	if len(permittedScopes.Projects) > 0 {
		query = fmt.Sprintf("%s or project_id in ?", query)
		values = append(values, permittedScopes.Projects)
	}
	return query, values
}

func CountApplications(
	ctx context.Context,
	tenantId string,
	permittedScopes *openapi.PermittedScopes,
	filterBy []validators.ParameterizedFilter,
	search string,
) (int64, error) {
	query := dao_utils.DB().WithContext(ctx).
		Model(&models.Application{}).
		Where(fmt.Sprintf("%s = ?", models.ApplicationColumnTenantId), tenantId)

	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		// if user is not system - filter by permitted scopes
		scopesCond, scopes := getPermissionQuery(permittedScopes)
		query = query.Where(scopesCond, scopes...)
	}

	query = dao_utils.ApplyParameterizedFilters(query, filterBy)

	if search != "" {
		query = dao_utils.ApplyTSSearchParam(query, search, "search_vector")
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count applications, err: %v", result.Error)
		return 0, result.Error
	}

	return count, nil
}

func GetApplicationById(ctx context.Context, tenantId string, applicationId uuid.UUID, permittedScopes *openapi.PermittedScopes) (*models.ApplicationView, error) {
	var application models.ApplicationView
	query := dao_utils.DB().WithContext(ctx).
		Model(&models.ApplicationView{}).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", applicationId)
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		// if user is not system - filter by permitted scopes
		scopesCond, scopes := getPermissionQuery(permittedScopes)
		query = query.Where(scopesCond, scopes...)
	}
	result := query.
		First(&application)

	if result.Error != nil {
		return nil, result.Error
	}

	return &application, nil
}

func CreateApplication(ctx context.Context, application *models.Application, spec *api.AiApplicationSpec, desiredPhase *string) *api.Error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin transaction for creating application %s: %v", application.Name, tx.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to create application %s", application.Name), nil).WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	if result := tx.Create(application); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create application %s, err %v", application.Name, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to create application %s", application.Name), nil).WithDetails(result.Error)
	}

	if spec != nil {
		syncEntry := &models.ApplicationCrudSync{
			ApplicationId: application.ID,
			TenantId:      application.TenantId,
			ClusterId:     application.ClusterId,
			Spec:          spec,
			DesiredPhase:  desiredPhase,
		}
		if result := tx.Create(syncEntry); result.Error != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create crud sync entry for application %s, err %v", application.Name, result.Error)
			return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to create application %s", application.Name), nil).WithDetails(result.Error)
		}
	}

	if result := tx.Commit(); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit transaction for creating application %s: %v", application.Name, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to create application %s", application.Name), nil).WithDetails(result.Error)
	}
	committed = true
	return nil
}

// MarkApplicationsAsDeleting transitions actual_phase to "deleting" for the given
// application IDs. Called when the agent pulls deletion records from application_crud_sync,
// because the agent is the component that actually performs the cluster-side deletion.
func MarkApplicationsAsDeleting(ctx context.Context, tenantId string, applicationIds []uuid.UUID, updatedBy string) *api.Error {
	if len(applicationIds) == 0 {
		return nil
	}
	result := dao_utils.DB().WithContext(ctx).
		Model(&models.Application{}).
		Where(&models.Application{TenantId: tenantId}).
		Where(fmt.Sprintf("%s IN ?", models.ApplicationColumnId), applicationIds).
		Where(fmt.Sprintf("%s != ? OR %s IS NULL", models.ApplicationColumnActualPhase, models.ApplicationColumnActualPhase), models.ApplicationPhaseDeleting).
		Where(fmt.Sprintf("%s != ?", models.ApplicationColumnActualPhase), models.ApplicationPhaseDeleted).
		Updates(map[string]interface{}{
			models.ApplicationColumnActualPhase: models.ApplicationPhaseDeleting,
			models.ApplicationColumnUpdatedAt:   time.Now(),
			models.ApplicationColumnUpdatedBy:   updatedBy,
		})

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to mark applications as deleting, err: %v", result.Error)
		return api.NewError(http.StatusServiceUnavailable, "Failed to mark applications as deleting", nil).WithDetails(result.Error)
	}
	return nil
}

func ListApplicationUpdates(ctx context.Context, tenantId string, clusterId string, afterUpdate int64, bulkSize int) ([]models.ApplicationSyncUpdate, *api.Error) {
	var records []models.ApplicationSyncUpdate

	result := dao_utils.DB().WithContext(ctx).
		Table(fmt.Sprintf("%s AS s", models.ApplicationCrudSyncTableName)).
		Select("a.id, a.created_at, a.created_by, a.name, a.source, a.type, a.tenant_id, a.cluster_id, a.department_id, a.project_id, a.project_name, a.actual_phase, s.spec, s.desired_phase, s.update_count").
		Joins(fmt.Sprintf("JOIN %s AS a ON s.%s = a.%s AND (a.%s IS NULL OR s.%s = ?)",
			models.ApplicationTableName,
			models.ApplicationCrudSyncColumnAppId,
			models.ApplicationColumnId,
			models.ApplicationColumnDeletedAt,
			models.ApplicationColumnDesiredPhase),
			models.ApplicationPhaseDeleted).
		Where(fmt.Sprintf("s.%s = ?", models.ApplicationColumnTenantId), tenantId).
		Where(fmt.Sprintf("s.%s = ?", models.ApplicationColumnClusterId), clusterId).
		Where(fmt.Sprintf("s.%s > ?", models.ApplicationCrudSyncColumnUpdateCount), afterUpdate).
		Limit(bulkSize).
		Order(fmt.Sprintf("s.%s", models.ApplicationCrudSyncColumnUpdateCount)).
		Find(&records)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list application updates for cluster %s, err: %v", clusterId, result.Error)
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to list application updates for cluster %s", clusterId), nil).WithDetails(result.Error)
	}

	return records, nil
}

func GetApplicationsHighestUpdateCount(ctx context.Context) (int64, *api.Error) {
	var currentValue int64
	result := dao_utils.DB().WithContext(ctx).Table(models.ApplicationCrudSyncSequence).Select(models.ApplicationCrudSyncSequenceLastVal).Scan(&currentValue)
	if result.Error != nil {
		return 0, api.NewError(http.StatusServiceUnavailable, "failed to read highest application update count", result.Error)
	}
	return currentValue, nil
}

// UpdateApplicationSyncMessage updates the sync status message reported by the agent for an application.
func UpdateApplicationSyncMessage(ctx context.Context, tenantId string, clusterId string, applicationId string, message string) *api.Error {
	appUUID, err := uuid.Parse(applicationId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("invalid application id: %s", applicationId), err)
	}
	result := dao_utils.DB().WithContext(ctx).
		Model(&models.Application{}).
		Where(&models.Application{
			BaseModel: dao_utils.BaseModel{ID: appUUID},
			TenantId:  tenantId,
			ClusterId: clusterId,
		}).
		Update(models.ApplicationColumnMessage, message)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update sync message for application %s, err: %v", applicationId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update sync message for application %s", applicationId), nil).WithDetails(result.Error)
	}

	return nil
}

func UpdateCrudSyncDesiredPhase(ctx context.Context, tenantId string, applicationId uuid.UUID, desiredPhase string) *api.Error {
	result := dao_utils.DB().WithContext(ctx).
		Model(&models.ApplicationCrudSync{}).
		Where(&models.ApplicationCrudSync{
			ApplicationId: applicationId,
			TenantId:      tenantId,
		}).
		Update(models.ApplicationColumnDesiredPhase, desiredPhase)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update crud sync desired phase for application %s, err: %v", applicationId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to update crud sync desired phase for application %s", applicationId), nil).WithDetails(result.Error)
	}
	return nil
}

// DeleteApplicationImmediate soft-deletes the applications row and sets
// desired_phase to "deleted" in application_crud_sync, in a single transaction.
// Used when the application has never been synced to the cluster (actual_phase
// is still "creating"), so the user sees the app disappear immediately.
// The crud_sync row is kept (rather than removed) so the agent still receives
// the delete instruction on its next poll — this handles the race where the
// agent already picked up the create but hasn't reported back yet.
func DeleteApplicationImmediate(ctx context.Context, tenantId string, applicationId uuid.UUID) *api.Error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin transaction for immediate delete of application %s: %v", applicationId, tx.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete application %s", applicationId), nil).WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	baseModel := dao_utils.BaseModel{ID: applicationId}
	if result := tx.Where(&models.Application{TenantId: tenantId, BaseModel: baseModel}).Delete(&models.Application{}); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to soft-delete application %s: %v", applicationId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete application %s", applicationId), nil).WithDetails(result.Error)
	}

	if result := tx.Model(&models.ApplicationCrudSync{}).
		Where(&models.ApplicationCrudSync{ApplicationId: applicationId, TenantId: tenantId}).
		Update(models.ApplicationColumnDesiredPhase, models.ApplicationPhaseDeleted); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update crud sync desired phase for application %s: %v", applicationId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete application %s", applicationId), nil).WithDetails(result.Error)
	}

	if result := tx.Commit(); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit immediate delete for application %s: %v", applicationId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to delete application %s", applicationId), nil).WithDetails(result.Error)
	}
	committed = true
	return nil
}

func CleanupApplicationCrudSync(ctx context.Context, tenantId string, clusterId string, afterUpdate int64) *api.Error {
	result := dao_utils.DB().WithContext(ctx).
		Where(&models.ApplicationCrudSync{
			TenantId:     tenantId,
			ClusterId:    clusterId,
			DesiredPhase: convutils.Ref(models.ApplicationPhaseDeleted),
		}).
		Where(fmt.Sprintf("%s < ?", models.ApplicationCrudSyncColumnUpdateCount), afterUpdate).
		Delete(&models.ApplicationCrudSync{})

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to cleanup application crud sync for cluster %s, err: %v", clusterId, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to cleanup application crud sync for cluster %s", clusterId), nil).WithDetails(result.Error)
	}
	return nil
}
