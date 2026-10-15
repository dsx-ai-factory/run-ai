/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	"github.com/google/uuid"
)

const (
	defaultWorkloadActionsLimit = 50
	maxWorkloadActionsLimit     = 500
)

func DbInsertWorkloadActions(ctx context.Context, records []models.WorkloadActionHistory) *api.Error {
	if len(records) == 0 {
		return nil
	}
	_, err := runaidb.DB().WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&records)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to insert workload action history", err)
	}
	return nil
}

func DbGetWorkloadActions(ctx context.Context, workloadID uuid.UUID, offset, limit int32) ([]models.WorkloadActionHistory, bool, *api.Error) {
	records := []models.WorkloadActionHistory{}
	tenantID, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return records, false, apiErr
	}
	if offset < 0 || limit < 0 {
		return records, false, api.NewValidationError("offset and limit must not be negative")
	}
	if limit == 0 {
		limit = defaultWorkloadActionsLimit
	}
	if limit > maxWorkloadActionsLimit {
		limit = maxWorkloadActionsLimit
	}
	err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ? AND workload_id = ?", tenantID, workloadID).
		Order("accepted_at DESC, id DESC").
		Offset(int(offset)).
		Limit(int(limit + 1)).
		Find(&records)
	if err != nil {
		return records, false, api.NewError(http.StatusInternalServerError, "Failed to list workload actions", err)
	}
	hasNext := len(records) > int(limit)
	if hasNext {
		records = records[:limit]
	}
	return records, hasNext, nil
}

func DbDeleteWorkloadActionHistoryOfTenantId(ctx context.Context, tenantID string) *api.Error {
	_, err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&models.WorkloadActionHistory{})
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to delete tenant workload action history", err)
	}
	return nil
}

func DbDeleteWorkloadActionHistoryOfClusterId(ctx context.Context, tenantID, clusterID string) *api.Error {
	_, err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ? AND cluster_id = ?", tenantID, clusterID).
		Delete(&models.WorkloadActionHistory{})
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to delete cluster workload action history", err)
	}
	return nil
}
