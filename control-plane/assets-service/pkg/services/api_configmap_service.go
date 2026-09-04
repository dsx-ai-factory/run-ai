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

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/projects"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/status_info"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/usage_info"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

// configMapApiService is a service that implements the logic in the config map service
type configMapApiService struct {
	dbAccessor ConfigMapsDBAccessor
}

// NewConfigMapService creates a default api service
func NewConfigMapService(dbAccessor ConfigMapsDBAccessor) api.ConfigMapAPIServicer {
	return &configMapApiService{
		dbAccessor: dbAccessor,
	}
}

// CreateConfigMapAsset handles ConfigMap asset creation request
func (cms *configMapApiService) CreateConfigMapAsset(ctx context.Context, request api.ConfigMapCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_CONFIG_MAP
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	err := validateConfigMapScope(ctx, request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", err.Error())
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Failed creating a Config Map due to scoping", nil)), err
	}

	nameValidationQuery, validationError := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_CONFIG_MAP,
		nil,
	)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	if convutils.IsStringHasValue(request.Spec.ConfigMap) {
		validationError = validateConfigMapAssetVsClusterResource(ctx, &request.Meta, request.Spec.ConfigMap)
		if validationError != nil {
			return NewErrorResponse(validationError), nil
		}
	}

	result, actionError := cms.dbAccessor.CreateConfigMapAsset(ctx, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)
	return api.Response(http.StatusAccepted, result), nil
}

func validateConfigMapScope(ctx context.Context, request api.ConfigMapCreationRequest) error {
	// cluster and department scopes are supported only starting cluster version 2.19
	switch request.Meta.Scope {
	case api.SCOPE_CLUSTER, api.SCOPE_DEPARTMENT:
		scopeSupported, err := checker.NewClusterCapabilitiesChecker(*request.Meta.ClusterId).
			ValidateClusterCapability(ctx, checker.MinVersionOfClusterConfigMapWideScope)
		if err != nil {
			return fmt.Errorf("failed to check version compatibility for ConfigMap data source: %v", err)
		}

		if !scopeSupported {
			return fmt.Errorf("creating ConfigMap data source in %s scope is not supported for cluser version", request.Meta.Scope)
		}
	case api.SCOPE_PROJECT:
	default:
		return fmt.Errorf("creating ConfigMap data source in %s scope is not supported", request.Meta.Scope)
	}

	return nil
}

// DeleteConfigMapAssetById handles ConfigMap asset deletion request
func (cms *configMapApiService) DeleteConfigMapAssetById(ctx context.Context, assetId string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_CONFIG_MAP
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

	actionError := cms.dbAccessor.DeleteConfigMapAsset(ctx, assetMeta, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return NewDeletionResponse(), nil
}

// GetConfigMapAssetById returns a ConfigMap asset by it asset id
func (cms *configMapApiService) GetConfigMapAssetById(ctx context.Context, assetId string, usageInfo bool, complianceProject int32,
	complianceWorkloadType string, statusInfo bool, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_CONFIG_MAP

	asset, actionError := cms.dbAccessor.GetConfigMapAssetById(ctx, assetId)
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
				patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_CONFIG_MAP)
				if patchErr != nil {
					return NewErrorResponse(patchErr), nil
				}
			}
		}
	}

	return api.Response(http.StatusOK, asset), nil
}

// ListConfigMapAssets returns a batch of ConfigMap assets
func (cms *configMapApiService) ListConfigMapAssets(ctx context.Context, name string, scope string, projectId int32, departmentId string, clusterId string, usageInfo bool,
	complianceProject int32, complianceWorkloadType string, statusInfo bool, assetIds string, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_CONFIG_MAP

	queryParams, queryError := query_types.NewQueryParams(assetIds, name, scope, false, projectId, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := cms.dbAccessor.ListConfigMapAssets(ctx, queryParams)
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
			patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_CONFIG_MAP)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, result), nil
}

// UpdateConfigMapAssetById handles ConfigMap asset update request
func (cms *configMapApiService) UpdateConfigMapAssetById(ctx context.Context, assetId string, request api.ConfigMapUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_CONFIG_MAP
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

	nameValidationQuery, err := scope.GetNameValidationQuery(ctx, request.Meta.Name, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, assetMeta.ClusterId, assetMeta.Private, api.ASSETKIND_CONFIG_MAP, &assetId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := cms.dbAccessor.UpdateConfigMapAsset(ctx, assetId, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusAccepted, result), nil
}

func validateConfigMapAssetVsClusterResource(ctx context.Context, meta *api.AssetCreationRequest, resourceName *string) *api.Error {
	if resourceName == nil {
		return nil
	}

	scopeInfo := ScopeInfoOfCreateRequest(meta)

	resp, err := k8sObjectsTracker.Get().ListConfigMapsByScopeAndName(ctx, string(scopeInfo.Scope), scopeInfo.ScopeId, *resourceName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to inquire config map status: %v", err.Message)
		return api.NewError(http.StatusServiceUnavailable, "failed to inquire the status of the config map in the cluster", errors.New(err.Message))
	}

	if len(resp.ConfigMaps) == 0 {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("config map resource '%s' of %s scope could not be found in the cluster", *resourceName, scopeInfo.Scope), nil)
	}

	return nil
}
