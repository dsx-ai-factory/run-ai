/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/projects"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/usage_info"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// NFSApiService is a service that implements the logic in the NFS service
// This service should implement the business logic for every endpoint for the host path asset API.
type NFSApiService struct{}

// NewNFSApiService creates a default api service
func NewNFSApiService() *NFSApiService {
	return &NFSApiService{}
}

func (s *NFSApiService) CreateNfsAsset(ctx context.Context, request api.NfsCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_NFS
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_NFS,
		nil,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.CreateNFSAsset(ctx, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)
	return api.Response(http.StatusAccepted, result), nil
}

func (s *NFSApiService) DeleteNfsAssetById(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_NFS
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	validationError := validateNotBeingUsedByTemplate(ctx, id, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	actionError := dao.DeleteNFSAsset(ctx, assetMeta, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return NewDeletionResponse(), nil
}

func (s *NFSApiService) GetNfsAssetById(ctx context.Context, id string,
	usageInfo bool, complianceProject int32, complianceWorkloadType string, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_NFS

	asset, actionError := dao.GetNFSAssetById(ctx, id)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	permittedScopes, allowedProjectIds, authError := query_types.ApplyRbacRulesGetById(ctx, asset.Meta)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	if usageInfo || complianceProject != 0 || complianceWorkloadType != "" {
		patchingPointers := response_patching.LoadResponsePatchingPtrs(asset)

		if usageInfo {
			patchErr := usage_info.PatchUsageInfo(ctx, assetKind, patchingPointers, allowedProjectIds)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}

		if complianceProject != 0 && complianceWorkloadType != "" {
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(patchingPointers, permittedScopes)

			patcher, patchErr := flat_workloads.NewComplianceInfoPatcher(ctx, complianceProject, complianceWorkloadType, complianceReplicaType,
				patchingPointers, permittedScopes, assetClusterStatus)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}

			patchErr = patcher.PatchComplianceInfo(ctx)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}

			patchErr = usage_info.PatchUsageTimes(ctx, complianceProject, complianceWorkloadType, patchingPointers)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, asset), nil
}

func (s *NFSApiService) ListNfsAssets(ctx context.Context, nameParam string, scopeParam string, projectIdParam int32,
	departmentId string, clusterId string, usageInfo bool, complianceProject int32, complianceWorkloadType string, assetIds string, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_NFS

	queryParams, queryError := query_types.NewQueryParams(assetIds, nameParam, scopeParam, false, projectIdParam, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := dao.ListNFSAssets(ctx, queryParams)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	patchingPointers := response_patching.LoadResponsePatchingPtrs(result)

	projects.PatchProjectNames(ctx, patchingPointers)

	if usageInfo {
		patchErr := usage_info.PatchUsageInfo(ctx, assetKind, patchingPointers, queryParams.AllowedOrgUnits.AllowedProjects)
		if patchErr != nil {
			return NewErrorResponse(patchErr), nil
		}
	}

	if complianceProject != 0 && complianceWorkloadType != "" {
		assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(patchingPointers, permittedScopes)

		patcher, patchErr := flat_workloads.NewComplianceInfoPatcher(ctx, complianceProject, complianceWorkloadType, complianceReplicaType,
			patchingPointers, permittedScopes, assetClusterStatus)
		if patchErr != nil {
			return NewErrorResponse(patchErr), nil
		}

		patchErr = patcher.PatchComplianceInfo(ctx)
		if patchErr != nil {
			return NewErrorResponse(patchErr), nil
		}

		patchErr = usage_info.PatchUsageTimes(ctx, complianceProject, complianceWorkloadType, patchingPointers)
		if patchErr != nil {
			return NewErrorResponse(patchErr), nil
		}
	}

	return api.Response(http.StatusOK, result), nil
}

func (s *NFSApiService) UpdateNfsAssetById(ctx context.Context, id string, request api.NfsUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_NFS
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id).WithEntityName(request.Meta.Name)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_UPDATE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(ctx, request.Meta.Name, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, assetMeta.ClusterId, assetMeta.Private, api.ASSETKIND_NFS, &id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.UpdateNFSAsset(ctx, id, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusAccepted, result), nil
}
