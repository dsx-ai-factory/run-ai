/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"gorm.io/gorm/clause"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/common"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

// ConfigMapsDBAccessor is an interface to interact with db for ConfigMap access
type ConfigMapsDBAccessor interface {
	// ListConfigMaps queries the configmap table for all configmaps that match the filter
	ListConfigMaps(ctx context.Context, filter *models.ConfigMapFilters, options *models.ListOptions) ([]models.ConfigMap, *api.Error)
	// CreateConfigMap inserts a new configmap into the configmap table and return the record
	CreateConfigMap(ctx context.Context, configMap *models.ConfigMap) (*models.ConfigMap, *api.Error)
	// DeleteConfigMap deletes all the configmaps that match the filter, from the configmap table
	DeleteConfigMap(ctx context.Context, filters *models.ConfigMapFilters) *api.Error
	// ApplyConfigMaps applies a batch of ConfigMaps and return a map of the record id-configMap that were applied
	ApplyConfigMaps(ctx context.Context, filters *models.ConfigMapFilters, configMaps []api.ConfigMap) (map[string]api.ConfigMap, *api.Error)
}

// configMapsDBAccessor implements ConfigMapsDAOAccessor
type configMapsDBAccessor struct{}

// NewConfigMapsDBAccessor returns a ConfigMap DB accessor
func NewConfigMapsDBAccessor() ConfigMapsDBAccessor {
	return &configMapsDBAccessor{}
}

// ListConfigMaps queries the configmap table for all configmaps that match the filter
func (a *configMapsDBAccessor) ListConfigMaps(ctx context.Context, filter *models.ConfigMapFilters, options *models.ListOptions) ([]models.ConfigMap, *api.Error) {
	if options == nil {
		options = &models.ListOptions{}
	}

	var apiErr *api.Error
	filter.TenantId, apiErr = common.TenantIdFromContext(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	record := models.NewConfigMap(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	query, err := filter.Apply(ctx, db.Model(record), options.IncludeAncestors, options.IncludeDescendants)
	if err != nil {
		return nil, err
	}

	configMaps := &[]models.ConfigMap{}
	if err := query.Order(filter.Name).Find(&configMaps); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return nil, dbErr
	}

	return *configMaps, nil
}

// CreateConfigMap inserts a new configmap into the configmap table
func (a *configMapsDBAccessor) CreateConfigMap(ctx context.Context, configMap *models.ConfigMap) (*models.ConfigMap, *api.Error) {
	var apiErr *api.Error
	configMap.TenantId, apiErr = common.TenantIdFromContext(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	dbErr := api.NewError(http.StatusServiceUnavailable, configMap.ErrorMessage("insert"), nil)

	db := runaidb.DB().WithContext(ctx)

	// FirstOrCreate will return the first record that matches the condition or create a new record if no records match
	if _, err := db.FirstOrCreate(configMap, configMap); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return nil, dbErr
	}

	return configMap, nil
}

// DeleteConfigMap deletes all the configmaps that match the filter, from the configmap table
func (a *configMapsDBAccessor) DeleteConfigMap(ctx context.Context, filters *models.ConfigMapFilters) *api.Error {
	var apiErr *api.Error
	filters.TenantId, apiErr = common.TenantIdFromContext(ctx)
	if apiErr != nil {
		return apiErr
	}

	record := models.NewConfigMap(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB().WithContext(ctx)

	if _, err := db.Where(filters).Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return dbErr
	}

	return nil
}

// ApplyConfigMaps applies a batch of ConfigMaps and return a map of the record ids-configMaps that were applied
func (a *configMapsDBAccessor) ApplyConfigMaps(ctx context.Context, filters *models.ConfigMapFilters, configMaps []api.ConfigMap) (map[string]api.ConfigMap, *api.Error) {
	record := models.NewConfigMap(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("apply"), nil)

	db := runaidb.DB().WithContext(ctx)

	tx := db.Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// Keep track on ConfigMaps' name, later on delete obs ConfigMaps
	cmNames := make([]string, 0)
	// Keep track on the ids of the ConfigMaps that were applied to relate the issues with the ConfigMaps
	cmIdMap := make(map[string]api.ConfigMap, len(configMaps))

	for _, cm := range configMaps {
		toAdd := models.NewConfigMap(&record.ConfigMapFilters)
		toAdd.Name = &cm.Name
		cmNames = append(cmNames, cm.Name)

		// FirstOrCreate will return the first record that matches the condition or create a new record if no records match
		if _, err := tx.Clauses(clause.OnConflict{DoNothing: true}).FirstOrCreate(&toAdd, &toAdd); err != nil {
			dbErr = dbErr.WithDetails(err)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
			return nil, dbErr
		}
		cmIdMap[toAdd.ID.String()] = cm
	}

	// Delete all records that are not in the list. for example
	// if I have in the table configmaps cm-a cm-b cm-c and I receive from the sync cm-b cm-c cm-d
	// I need to upsert cm-c and cm-d, but I also not to remove cm-a (because it is no longer in the cluster)
	if _, err := tx.
		Where(filters).
		Where(fmt.Sprintf("%s not in ?", models.ConfigMapNameColumn), cmNames).
		Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return nil, dbErr
	}

	if err := tx.Commit(); err != nil {
		dbErr = dbErr.WithDetails(err)
		if dbErr.Code != http.StatusConflict {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
			return nil, dbErr
		}
	}
	committed = true

	return cmIdMap, nil
}

// DeleteProjectConfigMaps deletes all configmap that of a project
func DeleteProjectConfigMaps(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, openapi.SCOPETYPE_PROJECT, tenantId, projectId, &models.ConfigMap{})
}

// DeleteClusterConfigMaps deletes all configmap within a cluster
func DeleteClusterConfigMaps(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	return deleteClusterEntities(ctx, tenantId, clusterId, &models.ConfigMap{})
}

// DeleteTenantConfigMaps deletes all configmap within a tenant
func DeleteTenantConfigMaps(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.ConfigMap{})
}
