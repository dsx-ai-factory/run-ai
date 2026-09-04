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

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	"gorm.io/gorm"
)

// DbUpsertWorkloadsElements syncs structure elements for multiple workloads in a single transaction,
// filtering out failed workloads. Accepts a map so workloads with zero elements are also processed
// (triggering deletion of stale rows). Per-workload failures are logged and rolled back but don't
// fail the entire batch — elements are non-critical.
func DbUpsertWorkloadsElements(ctx context.Context, tenantID uint, clusterID string, elementsByWorkload map[uuid.UUID][]models.WorkloadElement, failedWorkloads []string) {
	failedWorkloadsMap := createMapFromSlice(failedWorkloads)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	for workloadID, elements := range elementsByWorkload {
		if _, failed := failedWorkloadsMap[workloadID.String()]; failed {
			continue
		}
		upsertDbErr := dbUpsertWorkloadElements(ctx, tx, tenantID, clusterID, workloadID, elements)
		if upsertDbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to sync structure elements for workload <%s>, error: %s",
				workloadID, upsertDbErr.Message)
			tx.RollbackTo("sp")
		} else {
			tx.SavePoint("sp")
		}
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workloads elements upsert in cluster <%s>: %v",
			clusterID, err)
	} else {
		committed = true
	}
}

// DbUpsertWorkloadElements syncs structure elements for a single workload.
// Elements are immutable — new ones are inserted, removed ones are deleted, existing ones are skipped.
func DbUpsertWorkloadElements(ctx context.Context, tenantID uint, clusterID string, workloadID uuid.UUID, elements []models.WorkloadElement) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	dbErr := dbUpsertWorkloadElements(ctx, tx, tenantID, clusterID, workloadID, elements)
	if dbErr != nil {
		return dbErr
	}

	err := tx.Commit()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit workload elements upsert for workload %s: %v", workloadID, err)
		dbErr = api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to commit upsert workload elements for workload %s", workloadID), nil)
		return dbErr.WithDetails(err)
	}
	committed = true

	return nil
}

// DbGetWorkloadElements returns stored structure elements for a workload, ordered by level then name.
func DbGetWorkloadElements(ctx context.Context, tenantID uint, workloadID uuid.UUID, offset int32, limit int32) ([]models.WorkloadElement, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get elements for workload %s", workloadID), nil)
	var elements []models.WorkloadElement
	query := runaidb.DB().WithContext(ctx).
		Table(models.WorkloadElementsTableName).
		Where("tenant_id = ?", tenantID).
		Where("workload_id = ?", workloadID).
		Order("level, name")
	if offset > 0 {
		query = query.Offset(int(offset))
	}
	if limit > 0 {
		query = query.Limit(int(limit))
	}

	err := query.Find(&elements)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get elements for workload %s: %v", workloadID, err)
		return nil, dbErr.WithDetails(err)
	}
	return elements, nil
}

// DbCountWorkloadElements returns the number of stored structure elements for a workload.
func DbCountWorkloadElements(ctx context.Context, tenantID uint, workloadID uuid.UUID) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to count elements for workload %s", workloadID), nil)
	var count int64
	err := runaidb.DB().WithContext(ctx).
		Table(models.WorkloadElementsTableName).
		Where("tenant_id = ?", tenantID).
		Where("workload_id = ?", workloadID).
		Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count elements for workload %s: %v", workloadID, err)
		return 0, dbErr.WithDetails(err)
	}
	return count, nil
}

// DbGetWorkloadElement returns a single structure element by its ID within a workload.
func DbGetWorkloadElement(ctx context.Context, tenantID uint, workloadID uuid.UUID, elementID string) (*models.WorkloadElement, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get element %s for workload %s", elementID, workloadID), nil)
	var element models.WorkloadElement
	err := runaidb.DB().WithContext(ctx).
		Table(models.WorkloadElementsTableName).
		Where("tenant_id = ?", tenantID).
		Where("workload_id = ?", workloadID).
		Where("id = ?", elementID).
		First(&element)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("Element %s not found for workload %s", elementID, workloadID), nil)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get element %s for workload %s: %v", elementID, workloadID, err)
		return nil, dbErr.WithDetails(err)
	}
	return &element, nil
}

// DbGetElementPods returns pods belonging to a specific element.
func DbGetElementPods(ctx context.Context, tenantID uint, workloadID uuid.UUID, elementID string) ([]models.Pod, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get pods for element %s", elementID), nil)
	var pods []models.Pod
	err := runaidb.DB().WithContext(ctx).
		Model(&models.Pod{}).
		Joins("INNER JOIN pod_elements ON pod_elements.pod_id = pods.id").
		Where("pods.tenant_id = ?", tenantID).
		Where("pods.workload_id = ?", workloadID).
		Where("pod_elements.element_id = ?", elementID).
		Find(&pods)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get pods for element %s: %v", elementID, err)
		return nil, dbErr.WithDetails(err)
	}
	return pods, nil
}

func dbUpsertWorkloadElements(ctx context.Context, tx orm.DB, tenantID uint, clusterID string, workloadID uuid.UUID, elements []models.WorkloadElement) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to upsert workload elements for workload %s", workloadID), nil)

	// Collect incoming element IDs
	incomingIDs := make([]string, 0, len(elements))
	for _, e := range elements {
		incomingIDs = append(incomingIDs, e.ID)
	}

	// Delete elements that are no longer present.
	// clusterID may be empty when the request-level ClusterId is not set (e.g. SyncWorkloads
	// scoped by project only) — skip the filter to match ConstructDeleteQueryForProjectCluster behavior.
	// workload_id (together with tenant_id for sure) should be enough of a UUID
	// to delete all elements of this workload only.
	deleteQuery := tx.Table(models.WorkloadElement{}.TableName()).
		Where("tenant_id = ?", tenantID).
		Where("workload_id = ?", workloadID)
	if clusterID != "" {
		deleteQuery = deleteQuery.Where("cluster_id = ?", clusterID)
	}
	if len(incomingIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", incomingIDs)
	}
	if _, err := deleteQuery.Delete(&models.WorkloadElement{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete removed workload elements for workload %s: %v",
			workloadID, err)
		return dbErr.WithDetails(err)
	}

	// Insert new elements (existing ones are skipped via ON CONFLICT DO NOTHING)
	if len(elements) > 0 {
		_, err := tx.Table(models.WorkloadElement{}.TableName()).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&elements)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to insert workload elements for workload %s: %v",
				workloadID, err)
			return dbErr.WithDetails(err)
		}
	}

	return nil
}
