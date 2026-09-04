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

// GetRegistryById retrieves a registry by its ID.
func GetRegistryById(ctx context.Context, registryId string) (*models.Registry, *api.Error) {
	registryUuid, uuidErr := uuid.Parse(registryId)
	if uuidErr != nil {
		return nil, api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	rec := models.Registry{ID: registryUuid}
	rowsAffected, err := buildRegistryQuery(ctx, &rec).Scan(&rec)
	if err != nil || rowsAffected == 0 {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get registry '%s'", registryId), nil).WithDetailsFromDbResult(rowsAffected, err)
	}

	return &rec, nil
}

// ListRegistries lists all registries for the given tenant.
// Includes registries that are not owned by the tenant (system scope - presets).
func ListRegistries(ctx context.Context) ([]models.Registry, *api.Error) {
	var registries []models.Registry

	_, err := buildRegistryQuery(ctx, &models.Registry{}).Scan(&registries)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to list registries", nil).WithDetails(err)
	}

	return registries, nil
}

func ListPresetRegistries(ctx context.Context) (map[string]*models.Registry, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list preset registries", nil)
	db := runaidb.DB().WithContext(ctx)

	registries := make([]models.Registry, 0)

	filter := models.Registry{
		ScopeType: string(api.SCOPETYPE_SYSTEM),
	}

	if err := db.Model(&filter).Where(&filter).Find(&registries); err != nil {
		return nil, dbErr.WithDetails(err)
	}

	nameToMeta := make(map[string]*models.Registry)

	for _, registry := range registries {
		nameToMeta[registry.Name] = &registry
	}

	return nameToMeta, nil
}

// ValidateRegistryNameAvailable checks if the given name is available for a registry.
// The name must be unique per tenant and also unique across system scope registries.
// If excludeRegistryId is provided, it will be excluded from the check (useful for updates).
func ValidateRegistryNameAvailable(ctx context.Context, name string, excludeRegistryId *uuid.UUID) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to check registry name availability", nil)
	tenantId := auth.TenantIdOf(ctx)
	db := runaidb.DB().WithContext(ctx)

	var count int64
	query := db.Model(&models.Registry{}).
		Where(fmt.Sprintf("%s = ?", models.RegistryNameColumn), name).
		Where(fmt.Sprintf("(%s = ? AND %s = ?) OR %s = ?", models.RegistryScopeTypeColumn, models.RegistryScopeIdColumn, models.RegistryScopeTypeColumn),
			string(api.SCOPETYPE_TENANT),
			strconv.Itoa(int(tenantId)),
			string(api.SCOPETYPE_SYSTEM),
		)
	if excludeRegistryId != nil {
		query = query.Where("id != ?", *excludeRegistryId)
	}

	if err := query.Count(&count); err != nil {
		return dbErr.WithDetails(err)
	}

	if count > 0 {
		return api.NewError(http.StatusConflict, fmt.Sprintf("registry with name '%s' already exists", name), nil)
	}

	return nil
}

func CreateRegistry(ctx context.Context, registryModel *models.Registry) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to create registry", nil)
	db := runaidb.DB().WithContext(ctx)

	if _, err := db.Create(registryModel); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func UpdateRegistry(ctx context.Context, registryModel *models.Registry) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to update registry", nil)
	db := runaidb.DB().WithContext(ctx)

	rec := models.Registry{
		ID: registryModel.ID,
	}

	if _, err := db.Model(rec).Where(rec).Save(registryModel); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

// DeleteRegistry deletes a registry by its ID.
func DeleteRegistry(ctx context.Context, registryId string) *api.Error {
	registryUuid, uuidErr := uuid.Parse(registryId)
	if uuidErr != nil {
		return api.NewValidationError(fmt.Sprintf("invalid uuid: %v", uuidErr))
	}

	dbErr := api.NewError(http.StatusInternalServerError, "Failed to delete registry", nil)
	db := runaidb.DB().WithContext(ctx)

	rec := models.Registry{ID: registryUuid}
	if err := buildRegistryQuery(ctx, &rec).First(&rec); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return api.NewError(http.StatusNotFound, fmt.Sprintf("registry '%s' not found", registryId), nil)
		}
		return dbErr.WithDetails(err)
	}

	if _, err := db.Delete(&rec); err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func buildRegistryQuery(ctx context.Context, registryModel *models.Registry) orm.DB {
	tenantId := auth.TenantIdOf(ctx)
	db := runaidb.DB().WithContext(ctx)

	result := db.Model(registryModel).Where(registryModel)
	result = result.Where(fmt.Sprintf("(%s IS NULL or %s = '' or %s = ?)", models.RegistryTenantIdColumn, models.RegistryTenantIdColumn, models.RegistryTenantIdColumn), strconv.Itoa(int(tenantId)))

	return result
}
