package dao

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
)

const (
	RevisionsTableName         = "revisions"
	RevisionsViewName          = "revisions_view"
	RevisionsChildrenTableName = "revisions_children"
)

func DbUpsertRevisionsChildren(ctx context.Context, revisionsChildren []models.RevisionChild, failedRevisions []string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create revisions children", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	failedRevisionsMap := createMapFromSlice(failedRevisions)

	for _, child := range revisionsChildren {
		if _, ok := failedRevisionsMap[child.RevisionID.String()]; ok {
			continue
		}

		result := tx.Table(RevisionsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if result.Error != nil {
			return dbErr.WithDetails(result.Error)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the revisions children: %v", result.Error)
		return dbErr.WithDetails(result.Error)
	}
	committed = true

	return nil
}

func DbUpsertRevisionsPodsChildren(ctx context.Context, revisionsChildren []models.RevisionChild, filterOutPods []string) *api.Error {
	childrenToCreate := []models.RevisionChild{}
	filterOutPodsMap := createMapFromSlice(filterOutPods)

	for _, child := range revisionsChildren {
		if _, ok := filterOutPodsMap[child.ChildID.String()]; ok {
			continue
		}

		childrenToCreate = append(childrenToCreate, child)
	}

	return DbUpsertRevisionsChildren(ctx, childrenToCreate, nil)
}

func omitRevisionsEmptyStaticFields(query *gorm.DB, revision models.Revision) *gorm.DB {
	if revision.ClusterID.String() == "" {
		query = query.Omit("cluster_id")
	}
	if revision.DepartmentID == "" {
		query = query.Omit("department_id")
	}
	if revision.ProjectID == "" {
		query = query.Omit("project_id")
	}

	return query
}

func DbSyncRevisions(ctx context.Context, tenantId uint, clusterId string, projectId string, revisions []models.Revision, to *uuid.UUID, from *uuid.UUID, excludingIds []string) ([]string, []string, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to sync revisions", nil)

	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to begin a session with the db: %v", tx.Error)
		return nil, nil, dbErr.WithDetails(tx.Error)
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	syncedRevisions := []string{}
	failedRevisions := []string{}
	for _, revision := range revisions {
		query := tx.Table(RevisionsTableName)

		query = omitRevisionsEmptyStaticFields(query, revision)

		result := query.Save(&revision)
		if result.Error != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save revision %s in project %s, err %v", revision.Name, revision.ProjectID, result.Error)
			failedRevisions = append(failedRevisions, revision.ID.String())
		} else {
			tx.SavePoint("sp")
			syncedRevisions = append(syncedRevisions, revision.ID.String())
		}

		excludingIds = append(excludingIds, revision.ID.String())
	}

	if projectId == "" {
		return nil, nil, api.NewError(http.StatusBadRequest, "Revisions sync without scope is not supported", nil)
	}

	// if clusterId is empty, we do not consider it in the deletion query
	deleteQuery := ConstructDeleteQueryForProjectCluster(tx, RevisionsTableName, tenantId, clusterId, projectId, from, to, excludingIds, "")

	result := deleteQuery.
		Delete(&models.Revision{})

	if result.Error != nil {
		return nil, nil, dbErr.WithDetails(result.Error)
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the revisions sync: %v", result.Error)
		return nil, nil, dbErr.WithDetails(result.Error)
	}
	committed = true

	return syncedRevisions, failedRevisions, nil
}

func DbGetWorkloadRevisions(ctx context.Context, tenantId uint, workloadId string, permittedScopes *openapi.PermittedScopes, deleted bool) (*[]models.RevisionView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list revisions", nil)

	tx := dao_utils.DB().WithContext(ctx)

	revisions := &[]models.RevisionView{}

	query := tx.
		Table(RevisionsViewName).
		Where("tenant_id = ?", tenantId).
		Where("workload_id = ?", workloadId)

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	result := query.
		Order("id DESC").
		Find(revisions)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list revisions, err: %v", result.Error)
		return &[]models.RevisionView{}, dbErr.WithDetails(result.Error)
	}

	return revisions, nil
}

func DbCountWorkloadRevisions(ctx context.Context, tenantId uint, workloadId string, permittedScopes *openapi.PermittedScopes, deleted bool) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count revisions", nil)

	tx := dao_utils.DB().WithContext(ctx)

	query := tx.
		Table(RevisionsViewName).
		Where("tenant_id = ?", tenantId).
		Where("workload_id = ?", workloadId)

	if permittedScopes != nil {
		if permittedScopes.System == nil || !*permittedScopes.System {
			query = query.Where(getPermittedScopesFilter(*permittedScopes))
		}
	}

	if deleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count revisions, err: %v", result.Error)
		return 0, dbErr.WithDetails(result.Error)
	}

	return count, nil
}

func DbGetRevisionChildren(ctx context.Context, revisionId string, childTypes *[]string) (*[]models.RevisionChild, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get children of revision %s with childTypes %v", revisionId, childTypes), nil)
	tx := dao_utils.DB().WithContext(ctx)

	children := &[]models.RevisionChild{}
	query := tx.
		Table(RevisionsChildrenTableName).
		Where("revision_id = ?", revisionId)

	if childTypes != nil {
		query = query.Where("child_type IN ?", *childTypes)
	}

	result := query.Find(children)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get children of revision with id %s with childTypes %v, err: %v",
			revisionId, childTypes, result.Error)
		return &[]models.RevisionChild{}, dbErr.WithDetails(result.Error)
	}
	return children, nil
}

func DbGetRevision(ctx context.Context, tenantId uint, revisionId string) (*models.RevisionView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get revision %s", revisionId), nil)

	revision := &models.RevisionView{}

	tx := dao_utils.DB().WithContext(ctx)

	query := tx.
		Table(RevisionsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", revisionId)

	result := query.First(revision)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get revision with id %s, err: %v", revisionId, result.Error)
		return &models.RevisionView{}, dbErr.WithDetails(result.Error)
	}

	return revision, nil
}

func DbDeleteRevision(ctx context.Context, tenantId uint, revisionId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete revision %s", revisionId), nil)

	result := dao_utils.DB().WithContext(ctx).
		Table(RevisionsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Revision{}, revisionId)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete revision with id %s, err: %v", revisionId, result.Error)
		return dbErr.WithDetails(result.Error)
	}

	return nil
}
