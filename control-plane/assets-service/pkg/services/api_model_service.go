/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

// ModelApiService is a service that implements the logic for the ModelAPIServicer
// This service should implement the business logic for every endpoint for the ModelApi API.
// Include any external packages or services that will be required by this service.
type ModelApiService struct{}

// NewModelApiService creates a default api service
func NewModelApiService() api.ModelAPIServicer {
	return &ModelApiService{}
}

// CreateModelAsset - Create model asset.
func (s *ModelApiService) CreateModelAsset(ctx context.Context, request api.ModelCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_MODEL
	audit.AuditFromCtx(ctx).WithEntityName(request.Meta.Name).WithEntityType(string(assetKind)).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	workloadAssets, err := validateModelAssets(ctx, request.Meta.Scope, request.Meta.ProjectId, request.Meta.DepartmentId, &request.Spec.Assets)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_MODEL,
		nil,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	// create the model asset (model table) with the models assets ids
	result, actionError := dao.CreateModelAsset(ctx, request, workloadAssets, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)
	return api.Response(http.StatusCreated, result), nil
}

// DeleteModelAssetById - Delete model asset.
func (s *ModelApiService) DeleteModelAssetById(ctx context.Context, modelId string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_MODEL
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(modelId)

	assetMeta, getError := dao.GetAssetMetaById(ctx, modelId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	validationError := validateNotBeingUsedByTemplate(ctx, modelId, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	actionError := dao.DeleteModelAssetById(ctx, modelId, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return NewDeletionResponse(), nil
}

// GetModelAssetById - Get model asset.
func (s *ModelApiService) GetModelAssetById(ctx context.Context, modelId string, usageInfo bool, complianceProject int32, complianceWorkloadType string, complianceReplicaType string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_MODEL

	// TODO: usageInfo, complyToProject, complyToWorkloadType will be used later for inference policy
	assetMeta, getError := dao.GetAssetMetaById(ctx, modelId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_READ, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	result, actionError := dao.GetModelAssetById(ctx, modelId)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return api.Response(http.StatusOK, result), nil
}

// ListModelAssets - List model assets.
func (s *ModelApiService) ListModelAssets(ctx context.Context, name string, scope string, projectId int32, departmentId string,
	clusterId string, usageInfo bool, complianceProject int32, copmlianceWorkloadType string, environmentId string, complianceReplicaTypr string,
) (api.ImplResponse, error) {
	queryParams, queryError := query_types.NewQueryParams("", name, scope, false, projectId, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	_, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := dao.ListModels(ctx, queryParams, environmentId)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return api.Response(http.StatusOK, result), nil
}

// UpdateModelAssetById - Update model asset.
func (s *ModelApiService) UpdateModelAssetById(ctx context.Context, modelId string, request api.ModelUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_MODEL
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(modelId).WithEntityName(request.Meta.Name)

	assetMeta, getError := dao.GetAssetMetaById(ctx, modelId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_UPDATE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	_, err := validateModelAssets(ctx, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, &request.Spec.Assets)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(ctx, request.Meta.Name, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, assetMeta.ClusterId, assetMeta.Private, api.ASSETKIND_MODEL, &modelId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.UpdateModelAssetById(ctx, modelId, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusAccepted, result), nil
}

func validateModelAssets(ctx context.Context, scope api.Scope, projectId *int32, departmentId *string, assetIds *api.AssetsIds) ([]dbmodel.AssetsContainedInWorkload, *api.Error) {
	if convutils.IsEmptyOrNull(&assetIds.Environment) {
		return nil, api.NewValidationError("required field 'Environment' is zero value.")
	}

	containedAssetIds, assetsError := dao.ValidateContainedAssets(ctx, convutils.Ref(auth.WorkloadTypeInference), scope, nil, projectId, departmentId, assetIds)
	if assetsError != nil {
		return nil, assetsError
	}

	// When loading environment variables from an asset there aren't credentials from run param
	workloadAssets, _, dbError := dao.ListWorkloadAssets(containedAssetIds, nil, runaidb.DB().WithContext(ctx))
	if dbError != nil {
		return nil, dbError
	}

	return workloadAssets, nil
}
