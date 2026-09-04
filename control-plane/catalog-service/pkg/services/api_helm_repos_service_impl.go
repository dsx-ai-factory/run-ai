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

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
)

func (s *helmReposAPIService) createHelmRepo(ctx context.Context, request api.HelmRepoCreationRequest, migrating bool) (api.HelmRepo, *api.Error) {
	var scopeParams *scope.ScopeInfo
	switch request.ScopeType {
	case api.SCOPETYPE_SYSTEM:
		if !migrating {
			return api.HelmRepo{}, api.NewValidationError("system scope is not supported for helm repos")
		}
		scopeParams = scope.NewSystemScope()
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.HelmRepo{}, api.NewValidationError("non-tenant scope is not supported for helm repos")
	}

	if !migrating {
		rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_CREATE, false)
		if rbacErr != nil {
			return api.HelmRepo{}, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
		}
	}

	if err := validateHelmRepoURL(request.Url); err != nil {
		return api.HelmRepo{}, api.NewValidationError(fmt.Sprintf("invalid helm repo URL: %v", err))
	}

	helmRepoModel, err := apiToModelHelmRepo(ctx, request, migrating)
	if err != nil {
		return api.HelmRepo{}, api.NewValidationError(err.Error())
	}

	if api.ScopeType(scopeParams.Scope) != api.SCOPETYPE_SYSTEM {
		apiErr := dao.ValidateHelmRepoNameAvailable(ctx, request.Name, nil)
		if apiErr != nil {
			return api.HelmRepo{}, apiErr
		}
	}

	apiErr := dao.CreateHelmRepo(ctx, helmRepoModel)
	if apiErr != nil {
		return api.HelmRepo{}, apiErr
	}

	return modelToApiHelmRepo(ctx, *helmRepoModel), nil
}

func apiToModelHelmRepo(ctx context.Context, request api.HelmRepoCreationRequest, migrating bool) (*models.HelmRepo, error) {
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

	var description *string
	if request.Description != "" {
		description = &request.Description
	}

	return &models.HelmRepo{
		ID:          uuid.New(),
		Name:        request.Name,
		Url:         request.Url,
		IsPrivate:   request.IsPrivate,
		Description: description,
		ScopeType:   string(request.ScopeType),
		ScopeId:     request.ScopeId,
		TenantId:    tenantId,
		CreatedBy:   convutils.Deref(subjectId, "unknown"),
		UpdatedBy:   convutils.Deref(subjectId, "unknown"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (s *helmReposAPIService) updateHelmRepo(ctx context.Context, helmRepoId string, request api.HelmRepoUpdateRequest) (api.HelmRepo, *api.Error) {
	helmRepoUuid, err := uuid.Parse(helmRepoId)
	if err != nil {
		return api.HelmRepo{}, api.NewValidationError(fmt.Sprintf("invalid helm repo id: %v", err))
	}

	existingHelmRepo, apiErr := dao.GetHelmRepoById(ctx, helmRepoId)
	if apiErr != nil {
		return api.HelmRepo{}, apiErr
	}

	if existingHelmRepo.ScopeType == string(api.SCOPETYPE_SYSTEM) {
		return api.HelmRepo{}, api.NewValidationError("cannot update system scope helm repos")
	}

	var scopeParams *scope.ScopeInfo
	switch api.ScopeType(existingHelmRepo.ScopeType) {
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.HelmRepo{}, api.NewValidationError("non-tenant scope is not supported for helm repos")
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_UPDATE, false)
	if rbacErr != nil {
		return api.HelmRepo{}, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	if request.Name != "" && request.Name != existingHelmRepo.Name {
		apiErr := dao.ValidateHelmRepoNameAvailable(ctx, request.Name, &helmRepoUuid)
		if apiErr != nil {
			return api.HelmRepo{}, apiErr
		}
		existingHelmRepo.Name = request.Name
	}

	if request.Url != "" {
		if err := validateHelmRepoURL(request.Url); err != nil {
			return api.HelmRepo{}, api.NewValidationError(fmt.Sprintf("invalid helm repo URL: %v", err))
		}
		existingHelmRepo.Url = request.Url
	}

	if request.IsPrivate != nil {
		existingHelmRepo.IsPrivate = *request.IsPrivate
	}

	subjectId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if subjectId == nil {
		return api.HelmRepo{}, api.NewValidationError("subject id is missing")
	}

	existingHelmRepo.UpdatedBy = *subjectId
	existingHelmRepo.UpdatedAt = time.Now()
	existingHelmRepo.ID = helmRepoUuid

	apiErr = dao.UpdateHelmRepo(ctx, existingHelmRepo)
	if apiErr != nil {
		return api.HelmRepo{}, apiErr
	}

	return modelToApiHelmRepo(ctx, *existingHelmRepo), nil
}

func (s *helmReposAPIService) deleteHelmRepo(ctx context.Context, helmRepoId string) *api.Error {
	_, err := uuid.Parse(helmRepoId)
	if err != nil {
		return api.NewValidationError(fmt.Sprintf("invalid helm repo id: %v", err))
	}

	existingHelmRepo, apiErr := dao.GetHelmRepoById(ctx, helmRepoId)
	if apiErr != nil {
		return apiErr
	}

	if existingHelmRepo.ScopeType == string(api.SCOPETYPE_SYSTEM) {
		return api.NewValidationError("cannot delete system scope helm repos")
	}

	var scopeParams *scope.ScopeInfo
	switch api.ScopeType(existingHelmRepo.ScopeType) {
	case api.SCOPETYPE_TENANT:
		scopeParams = scope.NewTenantScope()
	default:
		return api.NewValidationError("non-tenant scope is not supported for helm repos")
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, authorization.RESOURCETYPE_REGISTRIES, authorization.ACTION_DELETE, false)
	if rbacErr != nil {
		return api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	apiErr = dao.DeleteHelmRepo(ctx, helmRepoId)
	if apiErr != nil {
		return apiErr
	}

	return nil
}
