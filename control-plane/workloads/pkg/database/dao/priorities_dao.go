/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

const (
	UniqueViolationErr              = "23505"
	UniqueClusterPriorityConstraint = "priorities_tenant_unique"
	SystemScopeType                 = "system"
	TenantScopeType                 = "tenant"
	ClusterScopeType                = "cluster"
)

func GetPriorities(ctx context.Context, tenantId string, sortOrder string, sortBy string, filterBy []queryapi.Filter) (*[]models.Priority, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.PrioritiesTableName).
		Preload(models.PriorityStatusPreload)

	var err error
	query, err = getPrioritiesQuery(ctx, query, tenantId, sortOrder, sortBy, filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get priorities query, err: %v", err)
		return nil, err
	}

	priorities := &[]models.Priority{}
	err = query.Find(priorities)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workload priorities, err: %v", err)
		return priorities, err
	}

	return priorities, nil
}

func GetClusterPriorities(ctx context.Context, tenantId string, clusterId string, sortOrder string, sortBy string, filterBy []queryapi.Filter) (*[]models.PriorityClusterView, error) {
	query := runaidb.DB().WithContext(ctx).
		Table(models.PrioritiesClusterViewName).
		Where("cluster_id = ?", clusterId)

	var err error
	query, err = getPrioritiesQuery(ctx, query, tenantId, sortOrder, sortBy, filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get priorities query, err: %v", err)
		return nil, err
	}

	priorities := &[]models.PriorityClusterView{}
	err = query.Find(priorities)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workload cluster priorities for clusterId <%s>, err: %v", clusterId, err)
		return priorities, err
	}

	return priorities, nil
}

func getPrioritiesQuery(ctx context.Context, query orm.DB, tenantId string, sortOrder string, sortBy string, filterBy []queryapi.Filter) (orm.DB, error) {
	query = query.Where("tenant_id = ? or scope_type='system'", tenantId)

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if sortBy != "" {
		query = query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
	}
	query = query.Order("id DESC")

	return query, nil
}

func GetPrioritiesSync(ctx context.Context, tenantId string, source string) (*[]models.Priority, error) {

	priorities := &[]models.Priority{}

	query := runaidb.DB().WithContext(ctx).
		Table(models.PrioritiesTableName).
		Where("tenant_id = ? or scope_type='system'", tenantId).
		Where("source = ?", source).
		Order("id DESC").
		Unscoped()

	err := query.Find(priorities)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workloads, err: %v", err)
		return priorities, err
	}

	return priorities, nil
}

func GetPriorityByName(ctx context.Context, tenantId string, priorityName string) (*models.Priority, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get priority by name %s", priorityName), nil)

	var priority models.Priority
	err := runaidb.DB().WithContext(ctx).
		Model(&models.Priority{}).
		Where("deleted_at IS NULL").
		Where("name = ?", priorityName).
		Where(fmt.Sprintf("(scope_type = '%s' OR %s = ?)", api.SCOPETYPE_SYSTEM, TenantIdColumn), tenantId).
		First(&priority)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get priority by name %s, err %v", priorityName, err)
		return nil, dbErr.WithDetails(err)
	}

	return &priority, nil
}

func GetPriorityById(ctx context.Context, tenantId, priorityId string) (*models.Priority, *api.Error) {
	errMessage := fmt.Sprintf("Failed to get priority by id %s", priorityId)
	dbErr := api.NewError(http.StatusInternalServerError, errMessage, nil)

	var priority models.Priority
	err := runaidb.DB().WithContext(ctx).
		Model(&models.Priority{}).
		Where("deleted_at IS NULL").
		Where("id = ?", priorityId).
		Where(fmt.Sprintf("(scope_type = '%s' OR %s = ?)", api.SCOPETYPE_SYSTEM, TenantIdColumn), tenantId).
		First(&priority)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, err %v", errMessage, err)
		return nil, dbErr.WithDetails(err)
	}

	return &priority, nil
}

func saveStatus(tx orm.DB, clusterId uuid.UUID, priority models.Priority) (models.Priority, error) {
	if len(priority.Status) != 1 {
		return priority, fmt.Errorf("failed to save priority %s status err: status invalid: %v", priority.ID, priority.Status)
	}
	query := tx.Table(models.PriorityStatusesTableName)

	_, err := query.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "priority_id"}, {Name: "cluster_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"phase"}),
	}).Save(&models.PriorityStatus{
		PriorityID: priority.ID,
		ClusterId:  clusterId,
		Phase:      priority.Status[0].Phase,
	})
	return priority, err
}

func deleteClusterPriorityClasses(tx orm.DB, tenantId string, clusterId string, priorityNamesToDelete []string) error {
	// delete cluster status for priorities since they were synced back with "deleted" phase
	// we will delete the entire PC since its origin is the cluster
	statusDeleteQuery := `DELETE FROM priority_statuses ps
			USING priorities p
			WHERE ps.priority_id = p.id
			AND p.name IN ?
			AND p.scope_id = ?
			AND p.scope_type = ?
			AND p.tenant_id = ?`

	_, err := tx.Exec(statusDeleteQuery, priorityNamesToDelete, clusterId, ClusterScopeType, tenantId)
	if err != nil {
		return err
	}

	priorityDeleteQuery := `DELETE FROM priorities p
			WHERE p.name IN ?
			AND p.scope_id = ?
			AND p.scope_type = ?
			AND p.tenant_id = ?`

	_, err = tx.Exec(priorityDeleteQuery, priorityNamesToDelete, clusterId, ClusterScopeType, tenantId)
	return err
}

func deleteRunaiPriorityClassesStatus(tx orm.DB, clusterId string, priorityIdsToDelete []string) error {
	// delete cluster status for priorities since they were synced back with "deleted" phase
	//
	statusDeleteQuery := `DELETE FROM priority_statuses ps
			USING priorities p
			WHERE ps.priority_id in ?
			AND ps.cluster_id = ?
			AND p.scope_type IN ?`
	_, err := tx.Exec(statusDeleteQuery, priorityIdsToDelete, clusterId, []string{SystemScopeType, TenantScopeType})
	return err
}

func deleteRunaiPriorityClassesWithExcluded(tx orm.DB, clusterId string, prioritiesIdsToExclude []string) error {
	// in case of full sync, we need to delete all PCs which are not received from cluster
	var err error
	statusDeleteQuery := `DELETE FROM priority_statuses ps
			USING priorities p
			WHERE ps.priority_id = p.id
			AND ps.cluster_id = ?
			AND p.scope_type IN ?`

	if len(prioritiesIdsToExclude) > 0 {
		statusDeleteQuery = statusDeleteQuery + ` AND p.id NOT IN ?`
		_, err = tx.Exec(statusDeleteQuery, clusterId, []string{SystemScopeType, TenantScopeType}, prioritiesIdsToExclude)
	} else {
		_, err = tx.Exec(statusDeleteQuery, clusterId, []string{SystemScopeType, TenantScopeType})
	}
	return err
}

func deleteClusterPriorityClassesWithExcluded(tx orm.DB, tenantId string, clusterId string, prioritiesToExclude []string) error {
	// in case of full sync, we need to delete all PCs which are not received from cluster
	var err error

	statusDeleteQuery := `DELETE FROM priority_statuses ps
			USING priorities p
			WHERE ps.priority_id = p.id
			AND ps.cluster_id = ?
			AND p.scope_id = ?
			AND p.scope_type = ?
			AND p.tenant_id = ?`

	priorityDeleteQuery := `DELETE FROM priorities p
			WHERE p.scope_id = ?
			AND p.scope_type = ?
			AND p.tenant_id = ?`

	if len(prioritiesToExclude) > 0 {
		statusDeleteQuery = statusDeleteQuery + ` AND p.name NOT IN ?`
		priorityDeleteQuery = priorityDeleteQuery + ` AND p.name NOT IN ?`
		_, err = tx.Exec(statusDeleteQuery, clusterId, clusterId, ClusterScopeType, tenantId, prioritiesToExclude)
		if err != nil {
			return err
		}
		_, err = tx.Exec(priorityDeleteQuery, clusterId, ClusterScopeType, tenantId, prioritiesToExclude)
	} else {
		_, err = tx.Exec(statusDeleteQuery, clusterId, clusterId, ClusterScopeType, tenantId)
		if err != nil {
			return err
		}

		_, err = tx.Exec(priorityDeleteQuery, clusterId, ClusterScopeType, tenantId)
	}
	return err
}

// DeletePrioritiesOfTenantId Delete all priorities in a specific tenant
func DeletePrioritiesOfTenantId(ctx context.Context, tenantId string) error {
	// By default, GORM's Delete() method only performs a soft delete if the model has gorm.DeletedAt.
	// Using .Unscoped() ensures a hard delete, permanently removing the record from the database.
	_, err := runaidb.DB().WithContext(ctx).Unscoped().
		Table(models.PrioritiesTableName).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Priority{})
	if err != nil {
		return fmt.Errorf("failed to workload delete priority in tenant %s. err: %w", tenantId, err)
	}
	return nil
}

// DeletePrioritiesOfClusterId Delete all priorities in a specific tenant
func DeletePrioritiesOfClusterId(ctx context.Context, tenantId string, clusterId string) error {
	_, err := runaidb.DB().WithContext(ctx).
		Table(models.PrioritiesTableName).
		Where("tenant_id = ?", tenantId).
		Where("scope_type = ?", ClusterScopeType).
		Where("scope_id = ?", clusterId).
		Delete(&models.Priority{})
	if err != nil {
		return fmt.Errorf("failed to workload delete priority in tenant %s. err: %w", tenantId, err)
	}
	return nil
}

// SyncClusterPrioritiesStatus Sync priorities which are not runai managed - created in cluster and synced to control-plane
func SyncClusterPrioritiesStatus(
	ctx context.Context,
	tenantId string,
	clusterId uuid.UUID,
	clusterPriorities []models.Priority,
	clusterPriorityNamesToDelete []string,
	syncDeleted bool,
) ([]string, []string, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	synced := []string{}
	failed := []string{}
	excludedNamesFromDelete := []string{}

	if len(clusterPriorities) > 0 {
		syncedExisting, failedExisting, names := handleClusterPriorityClasses(ctx, tx, clusterId, clusterPriorities)
		synced = append(synced, syncedExisting...)
		failed = append(failed, failedExisting...)
		excludedNamesFromDelete = append(excludedNamesFromDelete, names...)
	}
	if len(clusterPriorityNamesToDelete) > 0 {
		err := deleteClusterPriorityClasses(tx, tenantId, clusterId.String(), clusterPriorityNamesToDelete)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete priorities. err %v", err)
			failed = append(failed, clusterPriorityNamesToDelete...)
		} else {
			tx.SavePoint("sp")
			synced = append(synced, clusterPriorityNamesToDelete...)
		}
	}

	if syncDeleted {
		err := deleteClusterPriorityClassesWithExcluded(tx, tenantId, clusterId.String(), excludedNamesFromDelete)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete priorities full sync. err %v", err)
		} else {
			tx.SavePoint("sp")
		}
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit priorities sync: %v", err)
		return nil, nil, err
	}
	committed = true

	return synced, failed, nil
}

// SyncRunaiPrioritiesStatus Sync priorities which are runai managed - originated in control-plane and status synced back from cluster
func SyncRunaiPrioritiesStatus(
	ctx context.Context,
	clusterId uuid.UUID,
	priorities []models.Priority,
	deletedPriorityIds []string,
	syncDeleted bool,
) ([]string, []string, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	synced := []string{}
	failed := []string{}
	excludedIdsFromDelete := []string{}

	if len(priorities) > 0 {
		syncedExisting, failedExisting, idsExisting := handleRunaiPriorityClasses(ctx, tx, priorities, clusterId)
		synced = append(synced, syncedExisting...)
		failed = append(failed, failedExisting...)
		excludedIdsFromDelete = append(excludedIdsFromDelete, idsExisting...)
	}
	if len(deletedPriorityIds) > 0 {
		err := deleteRunaiPriorityClassesStatus(tx, clusterId.String(), deletedPriorityIds)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete priorities. err %v", err)
			failed = append(failed, deletedPriorityIds...)
		} else {
			synced = append(synced, deletedPriorityIds...)
			tx.SavePoint("sp")
		}
	}

	if syncDeleted {
		err := deleteRunaiPriorityClassesWithExcluded(tx, clusterId.String(), excludedIdsFromDelete)
		if err != nil {
			tx.RollbackTo("sp")
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete priorities full sync. err %v", err)
		} else {
			tx.SavePoint("sp")
		}
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit priorities sync: %v", err)
		return nil, nil, err
	}
	committed = true

	return synced, failed, nil
}

func handleRunaiPriorityClasses(ctx context.Context, tx orm.DB, existingPriorities []models.Priority, clusterId uuid.UUID) ([]string, []string, []string) {
	var err error
	synced := []string{}
	failed := []string{}
	ids := []string{}
	// save cluster status for each priority
	for _, priority := range existingPriorities {
		priority, err = saveStatus(tx, clusterId, priority)
		if err != nil {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save priority status %s. err %v", priority.ID.String(), err)
			failed = append(failed, priority.ID.String())
		} else {
			tx.SavePoint("sp") // save point
			synced = append(synced, priority.ID.String())
		}

		ids = append(ids, priority.ID.String())
	}
	return synced, failed, ids
}

func getPgError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

func handleClusterPriorityClasses(ctx context.Context, tx orm.DB, clusterId uuid.UUID, priorities []models.Priority) ([]string, []string, []string) {
	var err error
	synced := []string{}
	failed := []string{}
	names := []string{}
	for _, priority := range priorities {
		// priority is cluster originated - first create it, if it's already in db - update
		// (ON CONFLICT will not work in this case since "priorities_tenant_unique" db constraint is only partial)
		prioritySaved := false
		_, saveErr := tx.Table(models.PrioritiesTableName).Save(&priority)
		if saveErr != nil {
			pqErr := getPgError(saveErr)
			if pqErr != nil && pqErr.Code == UniqueViolationErr && pqErr.ConstraintName == UniqueClusterPriorityConstraint {
				// priority already exists - update
				tx.RollbackTo("sp") // Rollback to last saved point
				saveResultErr := tx.Raw(`UPDATE priorities SET 
                    value = ?, source = ? 
                  	WHERE name = ? AND scope_type = ? AND scope_id = ? AND tenant_id = ? RETURNING id`,
					priority.Value, priority.Source, priority.Name, priority.ScopeType, priority.ScopeId, priority.TenantId).Row().Scan(&priority.ID)

				if saveResultErr == nil {
					prioritySaved = true
				}
			}
		} else {
			prioritySaved = true
		}
		if prioritySaved {
			// if there was no error in upsert priority - save the status
			priority, err = saveStatus(tx, clusterId, priority)
			if err != nil {
				tx.RollbackTo("sp") // Rollback to last saved point
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save priority status %s. err %v", priority.ID.String(), err)
				failed = append(failed, priority.ID.String())
			} else {
				tx.SavePoint("sp") // save point
				synced = append(synced, priority.ID.String())
			}
		} else {
			tx.RollbackTo("sp") // Rollback to last saved point
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to save priority status %s. err %v", priority.ID.String(), err)
			failed = append(failed, priority.Name)
		}
		names = append(names, priority.Name)
	}
	return synced, failed, names
}
