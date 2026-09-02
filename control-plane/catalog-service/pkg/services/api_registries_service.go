/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// RegistriesAPIService is a service that implements the logic for the RegistriesAPIServicer
// This service should implement the business logic for every endpoint for the RegistriesAPI API.
// Include any external packages or services that will be required by this service.
type registriesAPIService struct {
}

// NewRegistriesAPIService creates a default api service
func NewRegistriesAPIService() *registriesAPIService {
	return &registriesAPIService{}
}

// ListRegistryIntegrations - List registry integrations
func (s *registriesAPIService) ListRegistryIntegrations(ctx context.Context) (api.ImplResponse, error) {
	registries, apiErr := dao.ListRegistries(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	registryIntegrations := make([]api.RegistryIntegration, len(registries))
	for i, registry := range registries {
		registryIntegrations[i] = modelToApiRegistryIntegration(ctx, registry)
	}

	return api.Response(http.StatusOK, api.RegistryIntegrationListResponse{
		Registries: registryIntegrations,
	}), nil
}

// GetRegistryIntegrationById - Get registry integration by id
func (s *registriesAPIService) GetRegistryIntegrationById(ctx context.Context, registryIntegrationId string) (api.ImplResponse, error) {
	registry, apiErr := dao.GetRegistryById(ctx, registryIntegrationId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	registryIntegration := modelToApiRegistryIntegration(ctx, *registry)

	return api.Response(http.StatusOK, registryIntegration), nil
}

// CreateRegistryIntegration - Create registry integration
func (s *registriesAPIService) CreateRegistryIntegration(ctx context.Context, registryIntegrationCreationRequest api.RegistryIntegrationCreationRequest) (api.ImplResponse, error) {
	registryIntegration, apiErr := s.createRegistry(ctx, registryIntegrationCreationRequest, false)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: create registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusOK, registryIntegration), nil
}

// UpdateRegistryIntegration - Update registry integration
func (s *registriesAPIService) UpdateRegistryIntegration(ctx context.Context, registryIntegrationId string, registryIntegrationUpdateRequest api.RegistryIntegrationUpdateRequest) (api.ImplResponse, error) {
	registryIntegration, apiErr := s.updateRegistry(ctx, registryIntegrationId, registryIntegrationUpdateRequest)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: update registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusOK, registryIntegration), nil
}

// DeleteRegistryIntegration - Delete registry integration
func (s *registriesAPIService) DeleteRegistryIntegration(ctx context.Context, registryIntegrationId string) (api.ImplResponse, error) {
	apiErr := s.deleteRegistry(ctx, registryIntegrationId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: delete registry: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusAccepted, nil), nil
}

func modelToApiRegistryIntegration(ctx context.Context, model models.Registry) api.RegistryIntegration {
	var tenantId int32
	if model.TenantId != nil {
		tenantIdInt, _ := strconv.ParseInt(*model.TenantId, 10, 32)
		tenantId = int32(tenantIdInt)
	} else {
		// In case of system registry, the tenant id is the tenant id of the user
		tenantId = auth.TenantIdOf(ctx)
	}

	return api.RegistryIntegration{
		Id:        model.ID.String(),
		ScopeType: api.ScopeType(model.ScopeType),
		ScopeId:   model.ScopeId,
		TenantId:  tenantId,
		Name:      model.Name,
		Url:       model.Url,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedBy: model.UpdatedBy,
	}
}
