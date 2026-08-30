package dao

import (
	"context"
	"net/http"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type IssuesChanges struct {
	Removed int64
	Added   int64
}

func ListResourceIssues(ctx context.Context, filter *models.ResourceIssueFilters) ([]models.ResourceIssue, *api.Error) {
	record := models.NewResourceIssue(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	issues := &[]models.ResourceIssue{}

	if err := db.Model(record).Where(filter).Order(filter.ResourceId).Find(&issues); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return *issues, nil
}

func ListResourceIssuesByIds(ctx context.Context, ids []string) ([]models.ResourceIssue, *api.Error) {
	record := models.NewResourceIssue(&models.ResourceIssueFilters{})
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	issues := []models.ResourceIssue{}

	if err := db.Model(record).Where(models.ResourceIdColumn+" IN ?", ids).Find(&issues); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return issues, nil
}

func UpdateResourceIssues(ctx context.Context, tenantId int32, clusterId string, resourceId string, issues []models.ResourceIssue) (*IssuesChanges, *api.Error) {
	tx := runaidb.DB().Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	deletedIssuesAmount, err := deleteResourceIssuesInTransaction(ctx, tx, tenantId, clusterId, resourceId)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "Failed to commit transaction", commitErr)
		}
		committed = true
		return &IssuesChanges{
			Added:   0,
			Removed: deletedIssuesAmount,
		}, nil
	}

	addedIssuesAmount, err := createResourceIssuesInTransaction(ctx, tx, issues)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "Failed to commit transaction", commitErr)
	}
	committed = true

	return &IssuesChanges{
		Removed: deletedIssuesAmount,
		Added:   addedIssuesAmount,
	}, nil
}

func createResourceIssuesInTransaction(ctx context.Context, tx orm.DB, records []models.ResourceIssue) (int64, *api.Error) {
	rowsAffected, err := tx.Create(records)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create resource issues, error: %v", err)
		return 0, api.NewError(http.StatusServiceUnavailable, "Failed to create resource issues", err)
	}

	return rowsAffected, nil
}

func deleteResourceIssuesInTransaction(ctx context.Context, tx orm.DB, tenantId int32, clusterId string, resourceIssueId string) (int64, *api.Error) {
	record := models.NewResourceIssue(&models.ResourceIssueFilters{ResourceId: &resourceIssueId})
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	rowsAffected, err := tx.Where("tenant_id = ?", tenantId).Where("cluster_id = ?", clusterId).Where("resource_id = ?", resourceIssueId).Delete(record)
	if err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
	}

	return rowsAffected, nil
}

// DeleteProjectResourceIssues deletes issues of resources within a project
func DeleteProjectResourceIssues(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, authorizationClient.SCOPETYPE_PROJECT, tenantId, projectId, &models.ResourceIssue{})
}

// DeleteClusterResourceIssues deletes issues of resources within a cluster
func DeleteClusterResourceIssues(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	return deleteClusterEntities(ctx, tenantId, clusterId, &models.ResourceIssue{})
}

// DeleteTenantResourceIssues deletes issues of resources within a tenant
func DeleteTenantResourceIssues(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.ConfigMap{})
}
