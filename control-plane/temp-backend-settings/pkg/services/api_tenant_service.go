/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authorization_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	bff_openapi "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client/generated"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	idm_openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	tm_client "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
	catalog_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
	dv_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	orgunit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	wc_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/autorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/adapters"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/providers"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/dao"
)

type BackendError struct {
	Status      int    `json:"status"`
	Description string `json:"description"`
	ErrorType   string `json:"errorType"`
}

func newBackendError(status int, description string) BackendError {
	return BackendError{
		Status:      status,
		Description: description,
		ErrorType:   "BackendError",
	}
}

type Clients struct {
	ClusterClient       cluster_service_client.IClusterClient
	ClusterAPIClient    *cluster_openapi.APIClient
	TenantMgrClient     tm_client.ITenantsManager
	IdmClient           *idm_openapi.APIClient
	BffClient           *bff_openapi.APIClient
	OrgUnitClient       *orgunit_openapi.APIClient
	CatalogClient       *catalog_openapi.APIClient
	DatavolumesClient   *dv_openapi.APIClient
	WorkloadsClient     *wc_openapi.APIClient
	AuthorizationClient *authorization_openapi.APIClient
}

type TenantAPIServiceDeps struct {
	*Clients
	SettingsDao    *dao.TenantSettingsDAO
	ShowAlphaFlags bool
}

type SettingsManagerInterface interface {
	GetAllSettings(ctx context.Context, tenantID int, tenantData settings.TenantStaticData, clusterVersions []string, excludeMigrated bool) ([]api.Setting, error)
	SetSetting(ctx context.Context, tenantID int, key string, value interface{}, clusterVersions []string) error
}

type TenantAPIService struct {
	settingsManager SettingsManagerInterface
	clusterClient   cluster_service_client.IClusterClient
	tenantMgrClient tm_client.ITenantsManager
}

func NewTenantAPIService(deps TenantAPIServiceDeps) *TenantAPIService {
	onSetDeps := buildOnSetDeps(deps)
	settingProviders := buildSettingProviders(deps)
	settingsManager := settings.NewSettingsManager(deps.SettingsDao, settingProviders, deps.ShowAlphaFlags, onSetDeps)

	return &TenantAPIService{
		settingsManager: settingsManager,
		clusterClient:   deps.ClusterClient,
		tenantMgrClient: deps.TenantMgrClient,
	}
}

func (s *TenantAPIService) GetTenantSettings(ctx context.Context, excludeMigrated bool, tenantIdParam string) (api.ImplResponse, error) {
	tenantID, errResp := s.resolveTenantID(ctx, tenantIdParam)
	if errResp != nil {
		return *errResp, nil
	}

	clusterVersions, err := s.getClusterVersions(ctx)
	if err != nil {
		logging.Errorf("failed to get cluster versions: %v", err)
		clusterVersions = []string{}
	}

	tenantData, err := s.getTenantStaticData(ctx, tenantID)
	if err != nil {
		logging.Errorf("failed to get tenant static data: %v", err)
		return api.Response(http.StatusInternalServerError, newBackendError(http.StatusInternalServerError, "failed to get tenant data")), nil
	}

	result, err := s.settingsManager.GetAllSettings(ctx, tenantID, tenantData, clusterVersions, excludeMigrated)
	if err != nil {
		logging.Errorf("failed to get tenant settings: %v", err)
		return api.Response(http.StatusInternalServerError, newBackendError(http.StatusInternalServerError, "failed to get tenant settings")), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (s *TenantAPIService) GetTenantSettingsByKey(ctx context.Context, settingKey string, tenantIdParam string) (api.ImplResponse, error) {
	tenantID, errResp := s.resolveTenantID(ctx, tenantIdParam)
	if errResp != nil {
		return *errResp, nil
	}

	tenantData, err := s.getTenantStaticData(ctx, tenantID)
	if err != nil {
		logging.Errorf("failed to get tenant static data: %v", err)
		return api.Response(http.StatusInternalServerError, newBackendError(http.StatusInternalServerError, "failed to get tenant data")), nil
	}

	allSettingsResult, err := s.settingsManager.GetAllSettings(ctx, tenantID, tenantData, []string{}, false)
	if err != nil {
		logging.Errorf("failed to get tenant settings: %v", err)
		return api.Response(http.StatusInternalServerError, newBackendError(http.StatusInternalServerError, "failed to get tenant settings")), nil
	}

	for _, setting := range allSettingsResult {
		if setting.Key == settingKey {
			return api.Response(http.StatusOK, setting.Value), nil
		}
	}

	return api.Response(http.StatusNotFound, newBackendError(http.StatusNotFound, fmt.Sprintf("setting %q not found", settingKey))), nil
}

func (s *TenantAPIService) UpdateTenantSetting(ctx context.Context, tenantIdParam string, req api.TenantSettingCreationRequest) (api.ImplResponse, error) {
	key := req.Key
	var value interface{}
	if req.Value != nil {
		value = *req.Value
	}
	tenantID, errResp := s.resolveTenantID(ctx, tenantIdParam)
	if errResp != nil {
		return *errResp, nil
	}

	clusterVersions, versErr := s.getClusterVersions(ctx)
	if versErr != nil {
		logging.Errorf("failed to get cluster versions: %v", versErr)
		clusterVersions = []string{}
	}

	if err := s.settingsManager.SetSetting(ctx, tenantID, key, value, clusterVersions); err != nil {
		logging.Errorf("failed to set setting %q: %v", key, err)
		var postSaveErr *settings.PostSaveError
		if errors.As(err, &postSaveErr) {
			return api.Response(http.StatusInternalServerError, newBackendError(http.StatusInternalServerError, err.Error())), nil
		}
		return api.Response(http.StatusBadRequest, newBackendError(http.StatusBadRequest, err.Error())), nil
	}

	return api.Response(http.StatusOK, api.TenantSettingCreationResponse{Msg: "done"}), nil
}

func (s *TenantAPIService) resolveTenantID(ctx context.Context, tenantIdParam string) (int, *api.ImplResponse) {
	tenantID, err := autorization.ResolveTenantID(ctx, tenantIdParam)
	if err != nil {
		logging.Errorf("failed to resolve tenant id: %v", err)
		resp := api.Response(http.StatusBadRequest, newBackendError(http.StatusBadRequest, err.Error()))
		return 0, &resp
	}
	return tenantID, nil
}

func (s *TenantAPIService) getClusterVersions(ctx context.Context) ([]string, error) {
	resp, err := s.clusterClient.GetClustersMinimalData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters minimal data: %w", err)
	}

	versions := make([]string, 0, len(resp.Clusters))
	for _, cluster := range resp.Clusters {
		v := cluster.Version.Get()
		if v != nil {
			versions = append(versions, *v)
		} else {
			versions = append(versions, "")
		}
	}
	return versions, nil
}

func (s *TenantAPIService) getTenantStaticData(ctx context.Context, tenantID int) (settings.TenantStaticData, error) {
	tenant, _, err := s.tenantMgrClient.GetTenantByIdInternal(ctx, int32(tenantID))
	if err != nil {
		return settings.TenantStaticData{}, fmt.Errorf("failed to get tenant by id: %w", err)
	}

	return settings.TenantStaticData{
		CreatedAt: tenant.CreatedAt,
	}, nil
}

func buildOnSetDeps(deps TenantAPIServiceDeps) *settings.OnSetDeps {
	orgUnitAdapter := &adapters.OrgUnitSettingsAdapter{
		OrgUnitAPI:  deps.OrgUnitClient.OrgUnitAPI,
		SettingsAPI: deps.OrgUnitClient.SettingsAPI,
	}
	idmAdapter := &adapters.IdentityManagerAdapter{
		SettingsAPI: deps.IdmClient.SettingsAPI,
	}
	bffWorkloadsAdapter := &adapters.BffWorkloadsAdapter{
		WorkloadsSettingsAPI: deps.BffClient.SettingsAPI,
	}
	bffNotificationsAdapter := &adapters.BffNotificationsAdapter{
		NotificationsSettingsAPI: deps.BffClient.SettingsAPI,
	}
	catalogAdapter := &adapters.CatalogAdapter{
		SettingsAPI: deps.CatalogClient.SettingsAPI,
	}
	datavolumesAdapter := &adapters.DatavolumesAdapter{
		SettingsAPI: deps.DatavolumesClient.SettingsAPI,
	}
	workloadsAdapter := &adapters.WorkloadsAdapter{
		SettingsAPI: deps.WorkloadsClient.SettingsAPI,
	}
	authorizationAdapter := &adapters.AuthorizationAdapter{
		SettingsAPI: deps.AuthorizationClient.SettingsAPI,
	}
	clusterAdapter := &adapters.ClusterAdapter{
		SettingsAPI: deps.ClusterAPIClient.SettingsAPI,
	}

	return &settings.OnSetDeps{
		OrgUnitSettingsUpdater:  orgUnitAdapter,
		IdentityManagerUpdater:  idmAdapter,
		BffWorkloadsUpdater:     bffWorkloadsAdapter,
		BffNotificationsUpdater: bffNotificationsAdapter,
		WorkloadsUpdater:        workloadsAdapter,
		CatalogUpdater:          catalogAdapter,
		DatavolumesUpdater:      datavolumesAdapter,
		AuthorizationUpdater:    authorizationAdapter,
		ClusterUpdater:          clusterAdapter,
		SettingsDao:             deps.SettingsDao,
	}
}

func buildSettingProviders(deps TenantAPIServiceDeps) []settings.SettingProvider {
	return []settings.SettingProvider{
		providers.NewIdentityManagerProvider(deps.IdmClient),
		providers.NewBffNotificationsProvider(deps.BffClient),
		providers.NewBffWorkloadsProvider(deps.BffClient),
		providers.NewCatalogProvider(deps.CatalogClient),
		providers.NewDatavolumesProvider(deps.DatavolumesClient),
		providers.NewWorkloadsProvider(deps.WorkloadsClient),
		providers.NewAuthorizationProvider(deps.AuthorizationClient),
		providers.NewClusterProvider(deps.ClusterAPIClient),
		providers.NewOrgUnitProvider(deps.OrgUnitClient),
	}
}
