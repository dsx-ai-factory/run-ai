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
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

func CreateApplicationChartUpload(ctx context.Context, upload *models.ApplicationChartUpload, filename string) *api.Error {
	if result := dao_utils.DB().WithContext(ctx).Create(upload); result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to store chart %q (id %s): %v", filename, upload.ID, result.Error)
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to store chart %q", filename), nil).WithDetails(result.Error)
	}
	return nil
}

// GetApplicationChartUploadByID looks up a stored chart by its id, scoped to the caller's tenant.
// The download endpoint is JWT-authenticated, so the token's tenant is the access boundary: a chart
// belonging to another tenant returns the same not-found as a missing id (no cross-tenant existence
// leak).
func GetApplicationChartUploadByID(ctx context.Context, id uuid.UUID, tenantId string) (*models.ApplicationChartUpload, *api.Error) {
	var upload models.ApplicationChartUpload
	// Both fields are always non-zero here (id is a parsed uuid, tenantId comes from the JWT), so
	// GORM's zero-value omission on struct conditions can't silently drop the tenant scope.
	result := dao_utils.DB().WithContext(ctx).
		Where(&models.ApplicationChartUpload{ID: id, TenantId: tenantId}).
		First(&upload)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("chart %s not found", id), nil)
		}
		return nil, api.NewError(http.StatusInternalServerError, "failed to get uploaded chart", nil).WithDetails(result.Error)
	}
	return &upload, nil
}
