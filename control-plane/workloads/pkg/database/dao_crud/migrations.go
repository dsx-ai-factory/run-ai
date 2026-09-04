/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"net/http"
)

func GetMigrationsInfo(ctx context.Context) (*models_crud.Migrations, *api.Error) {
	tenantId := int64(auth.TenantIdOf(ctx))

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get migration info of tenant '%d'", tenantId), nil)

	db := runaidb.DB().WithContext(ctx)

	rec := models_crud.Migrations{
		TenantId: tenantId,
	}

	rowsAffected, err := db.Model(rec).Where(rec).Scan(&rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, err)
		return nil, dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	if rowsAffected == 0 {
		return &models_crud.Migrations{
			TenantId:              tenantId,
			AssetsToFlatTemplates: 0,
		}, nil
	}

	return &rec, nil
}

func SetAssetsToFlatMigration(ctx context.Context, count int) *api.Error {
	tenantId := int64(auth.TenantIdOf(ctx))

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get migration info of tenant '%d'", tenantId), nil)

	db := runaidb.DB().WithContext(ctx)

	rec := models_crud.Migrations{
		TenantId: tenantId,
	}

	rowsAffected, err := db.Model(rec).Where(rec).Scan(&rec)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load: %v", dbErr.Message, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}

	rec.AssetsToFlatTemplates = int32(count)

	rowsAffected, err = db.Save(&rec)

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: save: %v", dbErr.Message, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}

	return nil
}
