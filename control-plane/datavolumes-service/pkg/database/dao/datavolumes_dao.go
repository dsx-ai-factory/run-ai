package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/services_utils"

	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/dbtables"

	"github.com/google/uuid"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"gorm.io/gorm/clause"
)

// DbGetDatavolumes -
// if usableInProjectIdAncestors is set - will get the datavolumes that are shared with (or originated in) the project id and its ancestors.
// else - will get the datavolumes that are originated in the permitted scopes.
func DbGetDatavolumes(ctx context.Context, tenantId uint,
	permittedScopes *authorization_api.PermittedScopes,
	usableInProjectIdAncestors *authorization_utils.Ancestors,
	offset int32, limit int32, sortBy string, sortOrder string, filterBy []queryapi.Filter, search string,
	includeInternal bool,
) ([]models.DatavolumeView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get datavolumes", nil)

	query := runaidb.DB().WithContext(ctx).
		Table(models.DatavolumeViewTableName).
		Where("tenant_id = ?", tenantId).
		Distinct(models.DatavolumeView{}.DistinctDatavolumeColumns())

	query, apiErr := applyUsableInProjectIdAncestors(ctx, query, permittedScopes, usableInProjectIdAncestors, dbErr)
	if apiErr != nil {
		return []models.DatavolumeView{}, apiErr
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	if !includeInternal {
		query.Where(fmt.Sprintf("%s != true", dbtables.DataVolumeColumnInternal))
	}

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	datavolumeViews := []models.DatavolumeView{}
	if err := query.Find(&datavolumeViews); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get datavolumes, err: %v", err)
		return []models.DatavolumeView{}, dbErr.WithDetails(err)
	}

	return datavolumeViews, nil
}

func DbGetDatavolumesInClusterWithPartialStatus(ctx context.Context, tenantId uint, clusterId uuid.UUID, offset int32, limit int32) ([]models.Datavolume, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get full datavolumes in cluster %s", clusterId.String()), nil)

	query := runaidb.DB().WithContext(ctx).
		Preload(models.SharedScopesPreloadKey).
		Table(models.DatavolumeTableName).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Omit(models.Datavolume{}.OmitStatusFieldsExceptPhase()...)

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}

	datavolumes := []models.Datavolume{}
	if err := query.Find(&datavolumes); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get datavolumes in cluster %s, err: %v", clusterId, err)
		return []models.Datavolume{}, dbErr.WithDetails(err)
	}

	return datavolumes, nil
}

func DbGetDatavolumeById(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, withSharedScopes bool) (*models.Datavolume, *api.Error) {
	return dbGetDatavolumeByIdInner(ctx, tenantId, datavolumeId, withSharedScopes, false)
}

func DbGetDatavolumeMeta(ctx context.Context, tenantId uint, datavolumeId uuid.UUID) (*models.Datavolume, *api.Error) {
	return dbGetDatavolumeByIdInner(ctx, tenantId, datavolumeId, false, true)
}

func dbGetDatavolumeByIdInner(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, withSharedScopes, metaOnly bool) (*models.Datavolume, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get datavolume %s", datavolumeId), nil)

	datavolume := &models.Datavolume{}
	query := runaidb.DB().WithContext(ctx).
		Table(models.DatavolumeTableName).
		Where("tenant_id = ?", tenantId)

	if withSharedScopes {
		query = query.Preload(models.SharedScopesPreloadKey)
	} else if metaOnly {
		query = query.Select(models.Datavolume{}.AncestorsAndPhaseFields())
	}

	if err := query.First(datavolume, datavolumeId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get datavolume with id %s, err: %v", datavolumeId, err)
		return nil, dbErr.WithDetails(err)
	}

	return datavolume, nil
}

func DbCreateDatavolume(ctx context.Context, datavolume models.Datavolume) (*models.Datavolume, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create datavolume", nil)

	if _, err := runaidb.DB().WithContext(ctx).Table(models.DatavolumeTableName).Create(&datavolume); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create datavolume %v, err: %v", datavolume, err)
		return nil, dbErr.WithDetails(err)
	}

	return &datavolume, nil
}

// DbCreateAndShareDatavolume for internal data volume creation, creating and sharing it on the scope it was created in.
// clusterName parameter is mandatory if the data volume scope is cluster
func DbCreateAndShareDatavolume(ctx context.Context, tenantId uint, datavolume models.Datavolume, clusterName string) (*models.Datavolume, *api.Error) {
	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		if _, err := tx.Table(models.DatavolumeTableName).Create(&datavolume); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create datavolume %v, err: %v", datavolume, err)
			return err
		}

		sharedScope := models.SharedScope{
			DatavolumeId: datavolume.ID,
			ScopeType:    datavolume.Scope,
			TenantId:     tenantId,
			CreatedBy:    datavolume.CreatedBy,
		}

		switch datavolume.Scope {
		case api.SCOPETYPE_DEPARTMENT:
			sharedScope.ScopeId = datavolume.DepartmentId
			sharedScope.DepartmentId = datavolume.DepartmentId
			sharedScope.ScopeName = datavolume.DepartmentName
		case api.SCOPETYPE_CLUSTER:
			sharedScope.ScopeId = datavolume.ClusterId.String()
			sharedScope.ScopeName = clusterName
		}

		if _, ssErr := tx.Table(models.SharedScopeTableName).
			Where("tenant_id = ?", tenantId).
			Where("datavolume_id = ?", datavolume.ID).
			Create(&sharedScope); ssErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to insert sharedscope %s/%s/%s for datavolume %s, err: %v",
				sharedScope.ScopeType, sharedScope.ScopeId, sharedScope.ScopeName, datavolume.ID, ssErr)
			// NOTE: preserves the pre-migration bug (see TestDbCreateAndShareDatavolume_ShareError):
			// the original returned the (nil) datavolume-create result here, swallowing this error
			// and committing the transaction. Fix is deferred to a follow-up PR.
			return nil
		}

		return nil
	})
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "Failed to create and share datavolume", err)
	}

	return &datavolume, nil
}

func DbDeleteDatavolume(ctx context.Context, tenantId uint, datavolumeId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete datavolume %s", datavolumeId), nil)

	// allow actually deleting the datavolume if its status prior to that was "Deleting";
	// or "Creating" - which means the DV was not created in the cluster yet.
	if _, err := runaidb.DB().WithContext(ctx).
		Table(models.DatavolumeTableName).
		Where("tenant_id = ?", tenantId).
		Where("phase IN ?", []string{string(api.DATAVOLUMEPHASE_DELETING), string(api.DATAVOLUMEPHASE_CREATING)}).
		Delete(&models.Datavolume{}, datavolumeId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete datavolume with id %s, err: %v", datavolumeId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

// DbUpdateStatus - should be used by cluster-sync; doesn't update updated_at, updated_by fields.
func DbUpdateStatus(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, status models.DatavolumeStatus) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update datavolume %s phase to %s", datavolumeId, status.Phase), nil)

	query := runaidb.DB().WithContext(ctx).
		Table(models.DatavolumeTableName).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", datavolumeId)

	if status.Phase != api.DATAVOLUMEPHASE_DELETING {
		// Don't update the datavolume if it's in the "Deleting" phase.
		// This is to prevent the cluster-sync from updating the datavolume phase before the agent pulls the deleting phase.
		query = query.Where("phase != ?", api.DATAVOLUMEPHASE_DELETING)
	}

	if _, err := updateStatusQuery(query, status); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update datavolume %s phase to: <%s/%s>, err: %v",
			datavolumeId, status.Phase, status.PhaseMessage, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

// DbUpdatePhaseToDeleting - used when the user sends a "DELETE" request.
// The actual deletion will be done after the cluster-sync sends an update about the deletion in the cluster.
func DbUpdatePhaseToDeleting(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, subjectId string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update datavolume %s phase to deleting", datavolumeId), nil)

	_, err := runaidb.DB().WithContext(ctx).
		Model(&models.Datavolume{}).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", datavolumeId).
		Updates(map[string]interface{}{"phase": api.DATAVOLUMEPHASE_DELETING, "phase_message": "", "updated_by": subjectId})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update datavolume %s phase to deleting, err: %v", datavolumeId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}

func DatavolumeNameAvailability(ctx context.Context, assetName string, clusterId string) (bool, *api.Error) {
	var count int64
	dbErr := api.NewError(
		http.StatusInternalServerError,
		fmt.Sprintf("Failed to check name availability for datavolume '%s' in cluster '%s'", assetName, clusterId),
		nil,
	)
	err := runaidb.DB().WithContext(ctx).
		Model(&models.Datavolume{}).
		Where("name = ?", assetName).
		Where("cluster_id = ?", clusterId).
		Count(&count)
	if err != nil {
		return false, dbErr.WithDetails(err)
	}

	return count == 0, nil
}

func GetUsedPvcsInPermittedScopes(ctx context.Context, projectId string) ([]string, *api.Error) {
	var pvcNames []string

	dbErr := api.NewError(
		http.StatusInternalServerError,
		fmt.Sprintf("Failed to fetch used PVCs in project '%s'", projectId),
		nil,
	)

	tenantId, _ := services_utils.ParseTenantId(ctx)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue GetUsedPvcsInPermittedScopes with nil permitted scopes. error: %v", err)
		permittedScopes = nil
	}

	scopeFilter := getPermittedScopesDbFilter(permittedScopes)

	query := runaidb.DB().WithContext(ctx).
		Model(&models.Datavolume{}).
		Where(models.Datavolume{
			TenantId:  tenantId,
			ProjectId: projectId,
		})

	if scopeFilter != "" {
		query = query.Where(scopeFilter)
	}

	if err := query.Pluck(models.DatavolumePvcNameColumn, &pvcNames); err != nil {
		return nil, dbErr.WithDetails(err)
	}

	return pvcNames, nil
}

func DbSyncDatavolumeStatuses(ctx context.Context, tenantId uint, clusterId string, datavolumeStatuses []models.Datavolume, from *uuid.UUID, to *uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to sync datavolumes in cluster %s", clusterId), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	updatedIds := []string{}
	for _, datavolume := range datavolumeStatuses {
		query := tx.Table(models.DatavolumeTableName).
			Where("tenant_id = ?", tenantId).
			Where("cluster_id = ?", clusterId).
			Where("id = ?", datavolume.ID)

		if datavolume.Phase != api.DATAVOLUMEPHASE_DELETING {
			// Don't update the datavolume if it's in the "Deleting" phase.
			// This is to prevent the cluster-sync from updating the datavolume phase before the agent pulls the deleting phase.
			query = query.Where("phase != ?", api.DATAVOLUMEPHASE_DELETING)
		}

		if _, err := updateStatusQuery(query, datavolume.DatavolumeStatus); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update datavolume %s phase as part of sync in cluster %s, err: %v",
				datavolume.ID, clusterId, err)
			return dbErr.WithDetails(err)
		}

		updatedIds = append(updatedIds, datavolume.ID.String())
	}

	deleteQuery := constructDeleteQueryInCluster(tx, models.DatavolumeTableName, tenantId, clusterId, from, to, updatedIds)
	if _, err := deleteQuery.Delete(&models.Datavolume{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete datavolumes as part of sync in cluster %s, err: %v",
			clusterId, err)
		return dbErr.WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the datavolumes sync in cluster %s, err: %v",
			clusterId, err)
		return dbErr.WithDetails(err)
	}
	committed = true
	return nil
}

func DbUpsertDatavolumeSharedScopes(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, toCreate, toDelete []models.SharedScope) ([]models.SharedScope, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert shared scopes for datavolume %s", datavolumeId), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	sortScopesByIdAndType(toDelete)
	for i := range toDelete {
		sharedScope := toDelete[i]
		_, err := tx.Table(models.SharedScopeTableName).
			Where("tenant_id = ?", tenantId).
			Where("datavolume_id = ?", datavolumeId).
			Where("scope_id = ?", sharedScope.ScopeId).
			Where("scope_type = ?", sharedScope.ScopeType).
			Delete(&models.SharedScope{})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete sharedscope %s/%s/%s for datavolume %s, err: %v",
				sharedScope.ScopeType, sharedScope.ScopeId, sharedScope.ScopeName, datavolumeId, err)
			return nil, dbErr.WithDetails(err)
		}
	}

	created := []models.SharedScope{}
	for i := range toCreate {
		sharedScope := toCreate[i]
		_, err := tx.Table(models.SharedScopeTableName).
			Where("tenant_id = ?", tenantId).
			Where("datavolume_id = ?", datavolumeId).
			Create(&sharedScope)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to insert sharedscope %s/%s/%s for datavolume %s, err: %v",
				sharedScope.ScopeType, sharedScope.ScopeId, sharedScope.ScopeName, datavolumeId, err)
			return nil, dbErr.WithDetails(err)
		}
		created = append(created, sharedScope)
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit upsert sharedscopes of datavolume %s, err: %v", datavolumeId, err)
		return nil, dbErr.WithDetails(err)
	}
	committedTx = true

	return created, nil
}

func DbPatchDatavolume(ctx context.Context, tenantId uint, datavolumeId uuid.UUID, subjectId, description string) (*models.Datavolume, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to patch datavolume %s", datavolumeId), nil)

	datavolume := &models.Datavolume{}
	_, err := runaidb.DB().WithContext(ctx).
		Model(datavolume).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", datavolumeId).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{"description": description, "updated_by": subjectId})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update datavolume %s description to: %s, err: %v",
			datavolumeId, description, err)
		return nil, dbErr.WithDetails(err)
	}

	return datavolume, nil
}

func DbCountDatavolumes(ctx context.Context,
	tenantId uint,
	permittedScopes *authorization_api.PermittedScopes,
	usableInProjectIdAncestors *authorization_utils.Ancestors,
	filterBy []queryapi.Filter,
	search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count datavolumes", nil)

	tx := runaidb.DB().WithContext(ctx)

	dataVolume := &models.DatavolumeView{}
	query := tx.Model(dataVolume).Where(fmt.Sprintf("%s = ?", models.DatavolumeTenantIdColumn), tenantId)

	query, apiErr := applyUsableInProjectIdAncestors(ctx, query, permittedScopes, usableInProjectIdAncestors, dbErr)
	if apiErr != nil {
		return 0, apiErr
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	// GORM can't use COUNT with DISTINCT so this is a workaround.
	distinctColumns := models.DatavolumeView{}.DistinctDatavolumeColumns()
	distinctExpr := fmt.Sprintf("COUNT(DISTINCT CONCAT(%s))", strings.Join(distinctColumns, ", ', ', "))

	var count int64
	_, err := query.
		Select(distinctExpr).
		Scan(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count distinct datavolumes, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func applyUsableInProjectIdAncestors(ctx context.Context,
	query orm.DB,
	permittedScopes *authorization_api.PermittedScopes,
	usableInProjectIdAncestors *authorization_utils.Ancestors,
	dbErr *api.Error) (orm.DB, *api.Error) {
	if usableInProjectIdAncestors != nil {
		usableInProjectIdDbFilter, err := getUsableInProjectIdDbFilter(usableInProjectIdAncestors)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count datavolumes shared with project id filter, err: %v", err)
			return query, dbErr.WithDetails(err)
		}
		query = query.Where(usableInProjectIdDbFilter)
	} else {
		// no need to apply the permitted scopes filter if the usableInProjectIdAncestors is set -
		// it should have already been validated that the requested usableInProjectId is permitted.
		scopesDbFilter := getPermittedScopesDbFilter(permittedScopes)
		if scopesDbFilter != "" {
			query = query.Where(scopesDbFilter)
		}
	}
	return query, nil
}
