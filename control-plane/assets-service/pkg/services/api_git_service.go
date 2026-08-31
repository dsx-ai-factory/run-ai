/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"

	scope2 "github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/projects"

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
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

// GitApiService is a service that implements the logic for the Git service
// This service should implement the business logic for every endpoint for the git asset API.
type GitApiService struct{}

// NewGitApiService creates a default api service
func NewGitApiService() *GitApiService {
	return &GitApiService{}
}

func (s *GitApiService) CreateGitAsset(ctx context.Context, request api.GitCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_GIT
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	err := validateGitAssetSpec(ctx, &request.Spec, request.Meta.ClusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(request.Meta.Scope), request.Meta.ProjectId, request.Meta.DepartmentId, request.Meta.ClusterId, nil)
	validationError = validateAssetVsCredentials(ctx, api.Scope(scopeInfo.Scope), scopeInfo.ScopeId, dbmodel.GetGitCredentialId(request.Spec))
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	nameValidationQuery, err := scope2.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_GIT,
		nil,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.CreateGitAsset(ctx, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)
	return api.Response(http.StatusAccepted, result), nil
}

func (s *GitApiService) DeleteGitAssetById(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_GIT
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	validationError := validateNotBeingUsedByTemplate(ctx, id, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	actionError := dao.DeleteGitAsset(ctx, assetMeta, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))
	return NewDeletionResponse(), nil
}

func (s *GitApiService) GetGitAssetById(ctx context.Context, id string,
	usageInfo bool, complianceProject int32, complianceWorkloadType string, statusInfo bool, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_GIT

	asset, actionError := dao.GetGitAssetById(ctx, id)
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
				patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_GIT)
				if patchErr != nil {
					return NewErrorResponse(patchErr), nil
				}
			}
		}
	}

	return api.Response(http.StatusOK, asset), nil
}

func (s *GitApiService) ListGitAssets(ctx context.Context, nameParam string, scopeParam string, projectIdParam int32,
	departmentId string, clusterId string, usageInfo bool, complianceProject int32, complianceWorkloadType string, statusInfo bool, assetIds string, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_GIT

	queryParams, queryError := query_types.NewQueryParams(assetIds, nameParam, scopeParam, false, projectIdParam, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := dao.ListGitAssets(ctx, queryParams)
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
			patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_GIT)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, result), nil
}

func (s *GitApiService) UpdateGitAssetById(ctx context.Context, id string, request api.GitUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_GIT
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

	err := validateGitAssetSpec(ctx, &request.Spec, assetMeta.ClusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nameValidationQuery, err := scope2.GetNameValidationQuery(ctx, request.Meta.Name, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, assetMeta.ClusterId, assetMeta.Private, api.ASSETKIND_GIT, &id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.UpdateGitAsset(ctx, id, request, validateAssetVsCredentials, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return api.Response(http.StatusAccepted, result), nil
}

func validateGitAssetSpec(ctx context.Context, gitSpec *api.GitAssetSpec, clusterId *string) *api.Error {
	if gitSpec.PasswordAssetId != nil && gitSpec.CredentialId != nil {
		return api.NewError(http.StatusBadRequest, "cannot provide both passwordAssetId and credentialId", nil)
	}
	// PasswordAssetId is deprecated, if provided move it to credentialId
	if gitSpec.PasswordAssetId != nil {
		gitSpec.CredentialId = gitSpec.PasswordAssetId
		gitSpec.PasswordAssetId = nil
	}

	if clusterId == nil {
		return nil
	}

	// new clusters since 2.15.0 support only one of branch or revision so we
	// validate that assets under these clusters have only one of them
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(*clusterId)
	clusterSupportsOneParam, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionGitSyncOneParameter)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if clusterSupportsOneParam {
		if gitSpec.Branch != nil && gitSpec.Revision != nil {
			return api.NewError(http.StatusBadRequest, `to point to your git revision, provide only one of the following: ‘branch’ OR ‘revision’`, nil)
		}
	}

	return nil
}
