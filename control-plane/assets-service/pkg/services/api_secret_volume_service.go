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

	scope2 "github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/projects"

	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/status_info"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/usage_info"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// SecretVolumeApiService is a service that implements the logic in the secret volume service
type SecretVolumeApiService struct {
	dbAccessor SecretVolumeDBAccessor
}

// NewSecretVolumeService creates a default api service
func NewSecretVolumeService(dbAccessor SecretVolumeDBAccessor) api.SecretAPIServicer {
	return &SecretVolumeApiService{
		dbAccessor: dbAccessor,
	}
}

// CreateSecretAsset handles Secret asset creation request
func (s *SecretVolumeApiService) CreateSecretAsset(ctx context.Context, request api.SecretAssetCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_SECRET_VOLUME
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	if request.Meta.Scope == api.SCOPE_TENANT {
		errStr := fmt.Sprintf("Creating Secret data source in %s scope isn't supported", request.Meta.Scope)
		return NewErrorResponse(api.NewError(http.StatusBadRequest, errStr, errors.New(errStr))), nil
	}

	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(request.Meta.Scope), request.Meta.ProjectId, request.Meta.DepartmentId, request.Meta.ClusterId, nil)
	validationError = validateAssetVsCredentials(ctx, request.Meta.Scope, scopeInfo.ScopeId, request.Spec.CredentialAssetId)
	if validationError != nil {
		// In case credential wasn't found, return bad request
		if validationError.Code == http.StatusNotFound {
			return api.Response(http.StatusBadRequest, fmt.Sprintf("Credential id wasn't found. credential id %s", *request.Spec.CredentialAssetId)), nil
		}
		return NewErrorResponse(validationError), nil
	}

	nameValidationQuery, validationError := scope2.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_SECRET_VOLUME,
		nil,
	)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	result, actionError := s.dbAccessor.CreateSecretVolumeAsset(ctx, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)
	return api.Response(http.StatusAccepted, result), nil
}

// DeleteSecretAssetById handles Secret asset deletion request
func (s *SecretVolumeApiService) DeleteSecretAssetById(ctx context.Context, assetId string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_SECRET_VOLUME
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(assetId)

	assetMeta, getError := dao.GetAssetMetaById(ctx, assetId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	validationError := validateNotBeingUsedByTemplate(ctx, assetId, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	actionError := s.dbAccessor.DeleteSecretVolumeAsset(ctx, assetMeta, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return NewDeletionResponse(), nil
}

// GetSecretAssetById returns a Secret asset by it asset id
func (s *SecretVolumeApiService) GetSecretAssetById(ctx context.Context, assetId string, usageInfo bool, complianceProject int32,
	complianceWorkloadType string, statusInfo bool, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_SECRET_VOLUME

	asset, actionError := s.dbAccessor.GetSecretVolumeAssetById(ctx, assetId)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	permittedScopes, allowedProjectIds, authError := query_types.ApplyRbacRulesGetById(ctx, asset.Meta)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	if usageInfo || statusInfo || complianceProject != 0 || complianceWorkloadType != "" {
		patchingPointers := response_patching.LoadResponsePatchingPtrs(asset)

		if usageInfo {
			patchErr := usage_info.PatchUsageInfo(ctx, assetKind, patchingPointers, allowedProjectIds)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}

		if (complianceProject != 0 && complianceWorkloadType != "") || statusInfo {
			assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(patchingPointers, permittedScopes)

			if complianceProject != 0 && complianceWorkloadType != "" {
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

			if statusInfo {
				patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_SECRET_VOLUME)
				if patchErr != nil {
					return NewErrorResponse(patchErr), nil
				}
			}
		}
	}

	return api.Response(http.StatusOK, asset), nil
}

// ListSecretAssets returns a batch of Secret assets
func (s *SecretVolumeApiService) ListSecretAssets(ctx context.Context, name string, scope string, projectId int32, departmentId string,
	clusterId string, usageInfo bool, complianceProject int32, complianceWorkloadType string, statusInfo bool, assetIds string,
	complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_SECRET_VOLUME

	queryParams, queryError := query_types.NewQueryParams(assetIds, name, scope, false, projectId, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := s.dbAccessor.ListSecretVolumeAssets(ctx, queryParams)
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

	if (complianceProject != 0 && complianceWorkloadType != "") || statusInfo {
		assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(patchingPointers, permittedScopes)

		if complianceProject != 0 && complianceWorkloadType != "" {
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

		if statusInfo {
			patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_SECRET_VOLUME)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, result), nil
}

// UpdateSecretAssetById handles secret volume asset update request
func (s *SecretVolumeApiService) UpdateSecretAssetById(ctx context.Context, assetId string, request api.SecretAssetUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_SECRET_VOLUME
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(assetId).WithEntityName(request.Meta.Name)

	assetMeta, getError := dao.GetAssetMetaById(ctx, assetId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_UPDATE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	nameValidationQuery, err := scope2.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		assetMeta.Scope,
		assetMeta.ProjectId,
		assetMeta.DepartmentId,
		assetMeta.ClusterId,
		assetMeta.Private,
		api.ASSETKIND_SECRET_VOLUME,
		&assetId,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := s.dbAccessor.UpdateSecretVolumeAsset(ctx, assetId, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusAccepted, result), nil
}
