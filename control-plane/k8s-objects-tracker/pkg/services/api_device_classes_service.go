/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

// DeviceClassesApiService implements the logic for the DeviceClassesApiServicer.
// A DeviceClass is cluster-scoped and, for tracking purposes, only its name is stored.
type DeviceClassesApiService struct{}

// NewDeviceClassesApiService creates a default api service
func NewDeviceClassesApiService() api.DeviceClassesAPIServicer {
	return &DeviceClassesApiService{}
}

// CreateDeviceClass - upsert a single device class observed in a cluster.
// Mirrors StorageClass per-item create: no pruning of other items in the scope.
func (s *DeviceClassesApiService) CreateDeviceClass(ctx context.Context, deviceClass api.DeviceClass) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a device class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	object, err := deviceClassApiToDb(tenantId, deviceClass)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a device class: %v", err)
		return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, "failed creating a device class", nil)), nil
	}

	if apiErr = dao.UpsertObjects(ctx, &object.ObjectFilters, []*models.Object{object}); apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a device class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed creating a device class", nil)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("created a device class")
	return api.Response(http.StatusOK, deviceClassDbToApi(object)), nil
}

// UpdateDeviceClass - update a single device class.
// Mirrors StorageClass: PUT and POST do the same work; the two verbs exist for
// client-side intent clarity.
func (s *DeviceClassesApiService) UpdateDeviceClass(ctx context.Context, deviceClass api.DeviceClass) (api.ImplResponse, error) {
	return s.CreateDeviceClass(ctx, deviceClass)
}

// DeleteDeviceClass - delete a single device class by name.
func (s *DeviceClassesApiService) DeleteDeviceClass(ctx context.Context, clusterId string, name string) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting a device class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format: %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filters := &models.ObjectFilters{
		Type:      models.ObjectTypeDeviceClass,
		Name:      &name,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}

	if apiErr = dao.Delete(ctx, filters); apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting a device class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed deleting a device class", nil)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("deleted a device class")
	return api.Response(http.StatusOK, nil), nil
}

// SyncDeviceClass - sync the full set of device classes observed in a cluster
func (s *DeviceClassesApiService) SyncDeviceClass(ctx context.Context, clusterId string, dcs api.DeviceClasses) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed syncing device classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filter := &models.ObjectFilters{
		Type:      models.ObjectTypeDeviceClass,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}

	records := deviceClassesApiToDb(tenantId, clusterUuid, dcs)

	// Note: A transaction is not needed here, since we wouldn't like to fail it all in case the deletion fails.
	apiErr = dao.UpsertObjects(ctx, filter, records)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed upserting device classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed syncing a device class", nil)), nil
	}

	apiErr = dao.DeleteObjectsNotInCluster(ctx, filter, records)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting device classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed syncing a device class", nil)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("synced device classes")
	return api.Response(http.StatusOK, dcs), nil
}

// GetDeviceClasses - get all device classes for a specific cluster
func (s *DeviceClassesApiService) GetDeviceClasses(ctx context.Context, clusterId string, name string) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting device class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, nil), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format: %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filters := &models.ObjectFilters{
		Type:      models.ObjectTypeDeviceClass,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}
	if name != "" {
		filters.Name = &name
	}

	objects, apiErr := dao.ListObjects(ctx, filters)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting device class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting device class list", nil)), nil
	}

	return api.Response(http.StatusOK, deviceClassesDbToApi(objects)), nil
}
