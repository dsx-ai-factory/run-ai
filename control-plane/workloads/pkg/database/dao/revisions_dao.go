/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
)

const (
	RevisionsTableName         = "revisions"
	RevisionsViewName          = "revisions_view"
	RevisionsChildrenTableName = "revisions_children"
)

func DbUpsertRevisionsChildren(ctx context.Context, revisionsChildren []models.RevisionChild, failedRevisions []string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create revisions children", nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	failedRevisionsMap := createMapFromSlice(failedRevisions)

	for _, child := range revisionsChildren {
		if _, ok := failedRevisionsMap[child.RevisionID.String()]; ok {
			continue
		}

		_, err := tx.Table(RevisionsChildrenTableName).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&child)
		if err != nil {
			return dbErr.WithDetails(err)
		}
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the revisions children: %v", err)
		return dbErr.WithDetails(err)
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

func omitRevisionsEmptyStaticFields(query orm.DB, revision models.Revision) orm.DB {
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

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	syncedRevisions := []string{}
	failedRevisions := []string{}
	for _, revision := range revisions {
		query := tx.Table(RevisionsTableName)

		query = omitRevisionsEmptyStaticFields(query, revision)

		_, err := query.Save(&revision)
		if err != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save revision %s in project %s, err %v", revision.Name, revision.ProjectID, err)
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

	_, err := deleteQuery.
		Delete(&models.Revision{})
	if err != nil {
		return nil, nil, dbErr.WithDetails(err)
	}

	err = tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the revisions sync: %v", err)
		return nil, nil, dbErr.WithDetails(err)
	}
	committed = true

	return syncedRevisions, failedRevisions, nil
}

func DbGetWorkloadRevisions(ctx context.Context, tenantId uint, workloadId string, permittedScopes *openapi.PermittedScopes, deleted bool) (*[]models.RevisionView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list revisions", nil)

	tx := runaidb.DB().WithContext(ctx)

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

	err := query.
		Order("id DESC").
		Find(revisions)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list revisions, err: %v", err)
		return &[]models.RevisionView{}, dbErr.WithDetails(err)
	}

	return revisions, nil
}

func DbCountWorkloadRevisions(ctx context.Context, tenantId uint, workloadId string, permittedScopes *openapi.PermittedScopes, deleted bool) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count revisions", nil)

	tx := runaidb.DB().WithContext(ctx)

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
	err := query.Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count revisions, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func DbGetRevisionChildren(ctx context.Context, revisionId string, childTypes *[]string) (*[]models.RevisionChild, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get children of revision %s with childTypes %v", revisionId, childTypes), nil)
	tx := runaidb.DB().WithContext(ctx)

	children := &[]models.RevisionChild{}
	query := tx.
		Table(RevisionsChildrenTableName).
		Where("revision_id = ?", revisionId)

	if childTypes != nil {
		query = query.Where("child_type IN ?", *childTypes)
	}

	err := query.Find(children)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get children of revision with id %s with childTypes %v, err: %v",
			revisionId, childTypes, err)
		return &[]models.RevisionChild{}, dbErr.WithDetails(err)
	}
	return children, nil
}

func DbGetRevision(ctx context.Context, tenantId uint, revisionId string) (*models.RevisionView, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get revision %s", revisionId), nil)

	revision := &models.RevisionView{}

	tx := runaidb.DB().WithContext(ctx)

	query := tx.
		Table(RevisionsViewName).
		Unscoped().
		Where("tenant_id = ?", tenantId).
		Where("id = ?", revisionId)

	err := query.First(revision)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get revision with id %s, err: %v", revisionId, err)
		return &models.RevisionView{}, dbErr.WithDetails(err)
	}

	return revision, nil
}

func DbDeleteRevision(ctx context.Context, tenantId uint, revisionId uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete revision %s", revisionId), nil)

	_, err := runaidb.DB().WithContext(ctx).
		Table(RevisionsTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Revision{}, revisionId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete revision with id %s, err: %v", revisionId, err)
		return dbErr.WithDetails(err)
	}

	return nil
}
