/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// StorageClassesApiService is a service that implements the logic for the StorageClassesApiServicer
type StorageClassesApiService struct{}

// NewStorageClassesApiService creates a default api service
func NewStorageClassesApiService() api.StorageClassesAPIServicer {
	return &StorageClassesApiService{}
}

const (
	RunaiNoneStorageClass = "runai-none"

	EmptyStructStr = "{}"
)

// CreateStorageClass - create storageClass
func (s *StorageClassesApiService) CreateStorageClass(ctx context.Context, storageClass api.StorageClass) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a storage class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	object, err := storageClassApiToDb(tenantId, storageClass)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a storage class: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed creating a storage class", nil)), nil
	}

	apiErr = dao.UpsertObjects(ctx, &object.ObjectFilters, []*models.Object{object})
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a storage class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed creating a storage class", nil)), nil
	}

	//
	//   whenever new storage class is discovered we attempt to create a default config for it.
	//
	defaultConfig := models.StorageClassConfiguration{
		StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{
			TenantId:  tenantId,
			ScopeType: convutils.Ref(string(openapi.SCOPETYPE_CLUSTER)),
			ScopeId:   convutils.Ref(storageClass.ClusterId),
			Name:      convutils.Ref(storageClass.StorageClassName),
		},
		Permissions: &api.StorageClassPermissions{
			AllowedForAssets:            convutils.Ref(true),
			AllowedForWorkloads:         convutils.Ref(true),
			AllowedForEphemeralVolumes:  convutils.Ref(true),
			AllowedForPersistentVolumes: convutils.Ref(true),
			AllowDataSharing:            convutils.Ref(false),
		},
	}

	_, configErr := dao.CreateStorageClassConfiguration(ctx, &defaultConfig)
	if configErr != nil && configErr.Code != http.StatusConflict {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create default config for storage class %v in cluster %v: %v",
			defaultConfig.Name, defaultConfig.ScopeId, configErr.Message)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("created a storage class")

	ret, err := storageClassDbToApi(object)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed creating a storage class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed creating a storage class", nil)), nil
	}
	return api.Response(http.StatusOK, ret), nil
}

func (s *StorageClassesApiService) UpdateStorageClass(ctx context.Context, storageClass api.StorageClass) (api.ImplResponse, error) {
	return s.CreateStorageClass(ctx, storageClass)
}

// SyncStorageClass - sync storageClasses
func (s *StorageClassesApiService) SyncStorageClass(ctx context.Context, clusterId string, scs api.StorageClasses) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed syncing storage classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filter := &models.ObjectFilters{
		Type:      models.ObjectTypeStorageClass,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}

	records, err := storageClassesApiToDb(tenantId, scs)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed upserting storage classes: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed syncing a storage class", nil)), nil
	}

	// Note: A transaction is not needed here, since we wouldn't like to fail it all in case the deletion fails.
	apiErr = dao.UpsertObjects(ctx, filter, records)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed upserting storage classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed syncing a storage class", nil)), nil
	}

	apiErr = dao.DeleteObjectsNotInCluster(ctx, filter, records)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting storage classes: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(int(apiErr.Code), "failed syncing a storage class", nil)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("synced storage classes")
	return api.Response(http.StatusOK, scs), nil
}

// DeleteStorageClass - delete storageClass
func (s *StorageClassesApiService) DeleteStorageClass(ctx context.Context, clusterId string, storageClassName string) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting a storage class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, apiErr), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format: %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filters := &models.ObjectFilters{
		Type:      models.ObjectTypeStorageClass,
		Name:      &storageClassName,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}

	apiErr = dao.Delete(ctx, filters)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting a storage class: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed deleting a storage class", nil)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("deleted a storage class")
	return api.Response(http.StatusOK, nil), nil
}

// GetStorageClasses - get all storageClasses for a specific cluster
func (s *StorageClassesApiService) GetStorageClasses(ctx context.Context, clusterId string, name string, includeNone bool) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, nil), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format: %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filters := &models.ObjectFilters{
		Type:      models.ObjectTypeStorageClass,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}
	if name != "" {
		filters.Name = &name
	}

	storageClasses, apiErr := dao.ListObjects(ctx, filters)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting storage class list", nil)), nil
	}

	if includeNone {
		runaiStorageClass := models.Object{
			ObjectFilters: *filters,
			// This isn't a real storage class, therefore there is no spec, but the parser expects a json.
			Spec: EmptyStructStr,
		}
		runaiStorageClass.Name = convutils.Ref(RunaiNoneStorageClass)
		storageClasses = append(storageClasses, &runaiStorageClass)
	}

	scs, err := storageClassesDbToApi(storageClasses)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting storage class list", nil)), nil
	}

	return api.Response(http.StatusOK, scs), nil
}

// GetStorageClasses - get all storageClasses for a specific cluster
func (s *StorageClassesApiService) GetStorageClassesV2(ctx context.Context, clusterId string, name string, includeNone bool) (api.ImplResponse, error) {
	tenantId, apiErr := common.TenantIdFromContext(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, nil), nil
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid cluster ID format: %v", err)
		return api.Response(http.StatusBadRequest, "invalid cluster ID format"), nil
	}

	filters := &models.ObjectFilters{
		Type:      models.ObjectTypeStorageClass,
		TenantId:  tenantId,
		ClusterId: clusterUuid,
		ScopeType: authclient.SCOPETYPE_CLUSTER,
	}
	if name != "" {
		filters.Name = &name
	}

	storageClasses, apiErr := dao.ListStorageClassWithOptionalConfig(ctx, filters)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", apiErr)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting storage class list", nil)), nil
	}

	if includeNone {
		runaiNoneConfig, err := dao.GetStorageClassConfigurationByName(ctx, filters.ClusterId.String(), RunaiNoneStorageClass, tenantId)
		if err != nil && err.Code != http.StatusNotFound {
			return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting storage configuration of 'runai-none'", errors.New(err.Message))), nil
		}

		runaiNoneInfo := models.StorageClassWithOptionalConfig{
			Object: models.Object{
				ObjectFilters: *filters,
				// This isn't a real storage class, therefore there is no spec, but the parser expects a json.
				Spec: EmptyStructStr,
			},
			StorageClassConfiguration: runaiNoneConfig,
		}

		runaiNoneInfo.ObjectFilters.Name = convutils.Ref(RunaiNoneStorageClass)

		storageClasses = append(storageClasses, &runaiNoneInfo)
	}

	scs, err := storageClassesDbToApiV2(storageClasses)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed getting storage class list: %v", err)
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "failed getting storage class list", nil)), nil
	}

	return api.Response(http.StatusOK, scs), nil
}

// V1CreateStorageClass - create a storage class [deprecated]
func (s *StorageClassesApiService) V1CreateStorageClass(ctx context.Context, clusterUuid string, v1SC api.V1StorageClass) (api.ImplResponse, error) {
	sc := v1StorageClassToApi(&v1SC, clusterUuid)
	resp, err := s.CreateStorageClass(ctx, *sc)
	if err != nil {
		return resp, err
	}
	return api.Response(resp.Code, storageClassToV1Api(resp.Body.(*api.StorageClass))), nil
}

// V1DeleteStorageClass - delete a storage class [deprecated]
func (s *StorageClassesApiService) V1DeleteStorageClass(ctx context.Context, clusterUuid string, storageClassName string) (api.ImplResponse, error) {
	return s.DeleteStorageClass(ctx, clusterUuid, storageClassName)
}

// V1GetStorageClasses - get all storageClasses from a specific cluster [deprecated]
func (s *StorageClassesApiService) V1GetStorageClasses(ctx context.Context, clusterUuid string, includeNone bool) (api.ImplResponse, error) {
	resp, err := s.GetStorageClasses(ctx, clusterUuid, "", includeNone)
	if err != nil {
		return resp, err
	}
	names := storageClassesToNameList(resp.Body.(*api.StorageClasses))
	return api.Response(resp.Code, api.V1GetStorageClasses200Response{StorageClasses: names}), nil
}

// V1SyncStorageClass - sync a storageClass [deprecated]
func (s *StorageClassesApiService) V1SyncStorageClass(ctx context.Context, uuid string, v1StorageClass []api.V1StorageClass) (api.ImplResponse, error) {
	scs := make([]api.StorageClass, len(v1StorageClass))
	for i, sc := range v1StorageClass {
		scs[i] = *v1StorageClassToApi(&sc, uuid)
	}
	return s.SyncStorageClass(ctx, uuid, api.StorageClasses{Items: scs})
}

// V1UpdateStorageClass - update a storage class [deprecated]
func (s *StorageClassesApiService) V1UpdateStorageClass(ctx context.Context, uuid string, storageClassName string, v1StorageClass api.V1StorageClass) (api.ImplResponse, error) {
	return s.V1CreateStorageClass(ctx, uuid, v1StorageClass)
}
