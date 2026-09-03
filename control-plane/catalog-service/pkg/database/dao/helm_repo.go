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
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
)

func GetHelmRepoById(ctx context.Context, helmRepoId string) (*models.HelmRepo, *api.Error) {
	helmRepoUuid, uuidErr := uuid.Parse(helmRepoId)
	if uuidErr != nil {
		return nil, api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	rec := models.HelmRepo{ID: helmRepoUuid}
	rowsAffected, err := buildHelmRepoQuery(ctx, &rec).Scan(&rec)
	if err != nil || rowsAffected == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("helm repo '%s' not found", helmRepoId), nil).WithDetailsFromDbResult(rowsAffected, err)
	}

	return &rec, nil
}

func ListHelmRepos(ctx context.Context) ([]models.HelmRepo, *api.Error) {
	var helmRepos []models.HelmRepo

	_, err := buildHelmRepoQuery(ctx, &models.HelmRepo{}).Scan(&helmRepos)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to list helm repos", nil).WithDetails(err)
	}

	return helmRepos, nil
}

func ListPresetHelmRepos(ctx context.Context) (map[string]*models.HelmRepo, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list preset helm repos", nil)
	db := runaidb.DB().WithContext(ctx)

	helmRepos := make([]models.HelmRepo, 0)

	filter := models.HelmRepo{
		ScopeType: string(api.SCOPETYPE_SYSTEM),
	}

	if err := db.Model(&filter).Where(&filter).Find(&helmRepos); err != nil {
		return nil, dbErr.WithDetails(err)
	}

	nameToMeta := make(map[string]*models.HelmRepo)

	for _, helmRepo := range helmRepos {
		nameToMeta[helmRepo.Name] = &helmRepo
	}

	return nameToMeta, nil
}

func ValidateHelmRepoNameAvailable(ctx context.Context, name string, excludeHelmRepoId *uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to check helm repo name availability", nil)
	tenantId := auth.TenantIdOf(ctx)
	db := runaidb.DB().WithContext(ctx)

	var count int64
	query := db.Model(&models.HelmRepo{}).
		Where(fmt.Sprintf("%s = ?", models.HelmRepoNameColumn), name).
		Where(fmt.Sprintf("(%s = ? AND %s = ?) OR %s = ?", models.HelmRepoScopeTypeColumn, models.HelmRepoScopeIdColumn, models.HelmRepoScopeTypeColumn),
			string(api.SCOPETYPE_TENANT),
			strconv.Itoa(int(tenantId)),
			string(api.SCOPETYPE_SYSTEM),
		)
	if excludeHelmRepoId != nil {
		query = query.Where("id != ?", *excludeHelmRepoId)
	}

	if err := query.Count(&count); err != nil {
		return dbErr.WithDetails(err)
	}

	if count > 0 {
		return api.NewError(http.StatusConflict, fmt.Sprintf("helm repo with name '%s' already exists", name), nil)
	}

	return nil
}

func CreateHelmRepo(ctx context.Context, helmRepoModel *models.HelmRepo) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create helm repo", nil)
	db := runaidb.DB().WithContext(ctx)

	if _, err := db.Create(helmRepoModel); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func UpdateHelmRepo(ctx context.Context, helmRepoModel *models.HelmRepo) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to update helm repo", nil)
	db := runaidb.DB().WithContext(ctx)

	rec := models.HelmRepo{
		ID: helmRepoModel.ID,
	}

	if _, err := db.Model(rec).Where(rec).Save(helmRepoModel); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func DeleteHelmRepo(ctx context.Context, helmRepoId string) *api.Error {
	helmRepoUuid, uuidErr := uuid.Parse(helmRepoId)
	if uuidErr != nil {
		return api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	dbErr := api.NewError(http.StatusInternalServerError, "Failed to delete helm repo", nil)
	db := runaidb.DB().WithContext(ctx)

	rec := models.HelmRepo{ID: helmRepoUuid}
	if err := buildHelmRepoQuery(ctx, &rec).First(&rec); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return api.NewError(http.StatusNotFound, fmt.Sprintf("helm repo '%s' not found", helmRepoId), nil)
		}
		return dbErr.WithDetails(err)
	}

	if _, err := db.Delete(&rec); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func buildHelmRepoQuery(ctx context.Context, helmRepoModel *models.HelmRepo) orm.DB {
	tenantId := auth.TenantIdOf(ctx)
	db := runaidb.DB().WithContext(ctx)

	result := db.Model(helmRepoModel).Where(helmRepoModel)
	result = result.Where(fmt.Sprintf("(%s IS NULL or %s = '' or %s = ?)", models.HelmRepoTenantIdColumn, models.HelmRepoTenantIdColumn, models.HelmRepoTenantIdColumn), strconv.Itoa(int(tenantId)))

	return result
}
