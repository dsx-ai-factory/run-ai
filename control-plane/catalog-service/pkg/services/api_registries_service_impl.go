/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
)

func (rs *registriesAPIService) createRegistry(ctx context.Context, request api.RegistryIntegrationCreationRequest, migrating bool) (api.RegistryIntegration, *api.Error) {
	var scopeParams *scope.ScopeInfo
	switch request.ScopeType {
	case api.SCOPETYPE_SYSTEM:
		if !migrating {
			return api.RegistryIntegration{}, api.NewValidationError("system scope is not supported for registry integrations")
		}
		scopeParams = scope.NewSystemScope()
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.RegistryIntegration{}, api.NewValidationError("non-tenant scope is not supported for registry integrations")
	}

	if !migrating {
		rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_CREATE, false)
		if rbacErr != nil {
			return api.RegistryIntegration{}, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
		}
	}

	registryModel, err := apiToModelRegistryIntegration(ctx, request, migrating)
	if err != nil {
		return api.RegistryIntegration{}, api.NewValidationError(err.Error())
	}

	if api.ScopeType(scopeParams.Scope) != api.SCOPETYPE_SYSTEM {
		apiErr := dao.ValidateRegistryNameAvailable(ctx, request.Name, nil)
		if apiErr != nil {
			return api.RegistryIntegration{}, apiErr
		}
	}

	apiErr := dao.CreateRegistry(ctx, registryModel)
	if apiErr != nil {
		return api.RegistryIntegration{}, apiErr
	}

	apiResponse := modelToApiRegistryIntegration(ctx, *registryModel)

	return apiResponse, nil
}

func apiToModelRegistryIntegration(ctx context.Context, request api.RegistryIntegrationCreationRequest, migrating bool) (*models.Registry, error) {
	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if !migrating && subjectId == nil {
		return nil, errors.New("subject id is missing")
	}

	var tenantId *string
	if !migrating {
		tenantId = convutils.Ref(strconv.Itoa(int(auth.TenantIdOf(ctx))))
	}

	if request.ScopeType == api.SCOPETYPE_TENANT && tenantId != nil {
		request.ScopeId = *tenantId
	}

	return &models.Registry{
		ID:        uuid.New(),
		Name:      request.Name,
		Url:       request.Url,
		ScopeType: string(request.ScopeType),
		ScopeId:   request.ScopeId,
		TenantId:  tenantId,
		CreatedBy: convutils.Deref(subjectId, "unknown"),
		UpdatedBy: convutils.Deref(subjectId, "unknown"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (rs *registriesAPIService) updateRegistry(ctx context.Context, registryId string, request api.RegistryIntegrationUpdateRequest) (api.RegistryIntegration, *api.Error) {
	registryUuid, err := uuid.Parse(registryId)
	if err != nil {
		return api.RegistryIntegration{}, api.NewValidationError(fmt.Sprintf("invalid registry id: %v", err))
	}

	existingRegistry, apiErr := dao.GetRegistryById(ctx, registryId)
	if apiErr != nil {
		return api.RegistryIntegration{}, apiErr
	}

	if existingRegistry.ScopeType == string(api.SCOPETYPE_SYSTEM) {
		return api.RegistryIntegration{}, api.NewValidationError("cannot update system scope registry integrations")
	}

	var scopeParams *scope.ScopeInfo
	switch api.ScopeType(existingRegistry.ScopeType) {
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.RegistryIntegration{}, api.NewValidationError("non-tenant scope is not supported for registry integrations")
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_UPDATE, false)
	if rbacErr != nil {
		return api.RegistryIntegration{}, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	if request.Name != "" && request.Name != existingRegistry.Name {
		apiErr := dao.ValidateRegistryNameAvailable(ctx, request.Name, &registryUuid)
		if apiErr != nil {
			return api.RegistryIntegration{}, apiErr
		}
		existingRegistry.Name = request.Name
	}

	if request.Url != "" {
		existingRegistry.Url = request.Url
	}

	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if subjectId == nil {
		return api.RegistryIntegration{}, api.NewValidationError("subject id is missing")
	}

	existingRegistry.UpdatedBy = *subjectId
	existingRegistry.UpdatedAt = time.Now()
	existingRegistry.ID = registryUuid

	apiErr = dao.UpdateRegistry(ctx, existingRegistry)
	if apiErr != nil {
		return api.RegistryIntegration{}, apiErr
	}

	apiResponse := modelToApiRegistryIntegration(ctx, *existingRegistry)

	return apiResponse, nil
}

func (rs *registriesAPIService) deleteRegistry(ctx context.Context, registryId string) *api.Error {
	_, err := uuid.Parse(registryId)
	if err != nil {
		return api.NewValidationError(fmt.Sprintf("invalid registry id: %v", err))
	}

	existingRegistry, apiErr := dao.GetRegistryById(ctx, registryId)
	if apiErr != nil {
		return apiErr
	}

	if existingRegistry.ScopeType == string(api.SCOPETYPE_SYSTEM) {
		return api.NewValidationError("cannot delete system scope registry integrations")
	}

	var scopeParams *scope.ScopeInfo
	switch api.ScopeType(existingRegistry.ScopeType) {
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.NewValidationError("non-tenant scope is not supported for registry integrations")
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_DELETE, false)
	if rbacErr != nil {
		return api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	apiErr = dao.DeleteRegistry(ctx, registryId)
	if apiErr != nil {
		return apiErr
	}

	return nil
}
