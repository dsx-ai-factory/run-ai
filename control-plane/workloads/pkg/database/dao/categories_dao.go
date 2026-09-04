/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	TenantIdColumn = "tenant_id"
)

// DbListCategories retrieves all categories for a given tenant and all system categories
func DbListCategories(ctx context.Context, tenantId uint) ([]api.WorkloadCategory, *api.Error) {
	var categories []api.WorkloadCategory
	query := runaidb.DB().WithContext(ctx).
		Model(&models.WorkloadsCategory{}).
		Where("deleted_at IS NULL").
		Where(fmt.Sprintf("(scope_type = '%s' OR %s = ?)", api.SCOPETYPE_SYSTEM, TenantIdColumn), tenantId)

	if err := query.Find(&categories); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list categories for tenant %d, err: %v", tenantId, err)
		return nil, api.NewError(503, "Failed to list categories", err)
	}

	return categories, nil
}

// DbGetCategoryByName gets a category by its name
func DbGetCategoryByName(ctx context.Context, tenantId uint, categoryName string) (*api.WorkloadCategory, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get category by name %s", categoryName), nil)

	var category api.WorkloadCategory
	err := runaidb.DB().WithContext(ctx).
		Model(&models.WorkloadsCategory{}).
		Where("deleted_at IS NULL").
		Where("name = ?", categoryName).
		Where(fmt.Sprintf("(%s = ? OR scope_type = '%s')", TenantIdColumn, api.SCOPETYPE_SYSTEM), tenantId).
		Order(fmt.Sprintf("CASE WHEN scope_type = '%s' THEN 1 ELSE 0 END ASC", api.SCOPETYPE_SYSTEM)).
		First(&category)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get category by name %s, err %v", categoryName, err)
		return nil, dbErr.WithDetails(err)
	}

	return &category, nil
}

// DbGetCategoryById retrieves a category by its ID
func DbGetCategoryById(ctx context.Context, tenantId uint, categoryId string) (*api.WorkloadCategory, *api.Error) {
	var category api.WorkloadCategory
	err := runaidb.DB().WithContext(ctx).
		Model(&models.WorkloadsCategory{}).
		Where("deleted_at IS NULL").
		Where("id = ?", categoryId).
		Where(fmt.Sprintf("(scope_type = '%s' OR %s = ?)", api.SCOPETYPE_SYSTEM, TenantIdColumn), tenantId).
		First(&category)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get category %s, err %v", categoryId, err)
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get category %s", categoryId), err)
	}

	return &category, nil
}
