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

	return api.Response(http.StatusOK, deviceClassesDbToApiV2(objects)), nil
}
