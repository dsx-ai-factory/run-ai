/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	datavolumes_service "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/datavolumes-service"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/projects"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/status_info"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/usage_info"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	scopeInfo "github.com/run-ai/runai/runai-application-packages/scopeinfo"
)

const (
	// dvFilterClusterId is the filter name for the cluster id property of the dv
	dvFilterClusterId = "clusterId"
)

// buildFilterFromQueryParams build filters from query params
func buildFiltersFromQueryParams(queryParams *query_types.QueryParams) []string {
	var filters []string

	if queryParams.ClusterId != nil {
		filters = append(filters, fmt.Sprintf("%s==%s", dvFilterClusterId, convutils.Deref(queryParams.ClusterId, "")))
	}

	return filters
}

func validatePvcNotBeingUsedByDatavolumes(
	ctx context.Context,
	pvcAsset *api.PvcAsset,
) *api.Error {
	filters := buildFiltersFromQueryParams(&query_types.QueryParams{
		ClusterId: pvcAsset.Meta.ClusterId,
	})
	allDatavolumes, err := datavolumes_service.Get().GetDataVolumesNotPaginated(
		ctx, openapi.DATAVOLUMEREQUESTTYPE_ORIGINATED, "", filters, "", "")
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to get datavolumes", err)
	}

	var datavolumeNamesUsingPvc []string
	claimName := pvcAsset.Spec.ClaimName
	if claimName == nil {
		return nil
	}
	for _, dv := range allDatavolumes {
		if dv.GetOriginPvcName() == *claimName {
			datavolumeNamesUsingPvc = append(datavolumeNamesUsingPvc, dv.Name)
		}
	}

	if len(datavolumeNamesUsingPvc) > 0 {
		return composeBeingInUseMessage("datavolumes", datavolumeNamesUsingPvc, api.ASSETKIND_PVC, pvcAsset.Meta.Id)
	}

	return nil
}

// agent supporting assets is not ready yet so temporarily avoid soft delete for assets
func agentSyncSupported(ctx context.Context, clusterId *string) bool {
	if clusterId == nil {
		return false
	}

	clusterChecker := checker.NewClusterCapabilitiesChecker(*clusterId)
	isSupported, err := clusterChecker.ValidateClusterCapability(ctx, checker.MinVersionOfAssetsAgentSync)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check capabilities of cluster %s, assume agent sync not supported", *clusterId)
		return false
	}

	return isSupported
}

// PVCApiService is a service that implements the logic in the PVC service
type PVCApiService struct{}

// NewPVCApiService creates a default api service
func NewPVCApiService() *PVCApiService {
	return &PVCApiService{}
}

func (s *PVCApiService) CreatePvcAsset(ctx context.Context, request api.PvcCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), authClient.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	var meta dbmodel.AssetMeta
	copyErr := copier.Copy(&meta, &request.Meta)
	if copyErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to load asset meta", copyErr)), nil
	}

	validationError = validatePvcSpec(ctx, &meta, &request.Spec, false)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	validationError = validateEphemeralAutoDelete(&request.Spec, &request.Meta)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	var dvId *string
	if convutils.BoolIsTrue(request.Spec.DataSharing) {
		var err *api.Error
		dvId, err = createDatavolumeForSharedPvc(ctx, request)
		if err != nil {
			return NewErrorResponse(err), nil
		}
	}

	nameValidationQuery, apiErr := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_PVC,
		nil,
	)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	// New PVC (not based on existing claim) can't use an already existing claim
	var claimNameValidationQuery *database.ParametrizedQuery
	if convutils.BoolIsFalse(request.Spec.ExistingPvc) && request.Spec.ClaimName != nil {
		query, apiErr := scope.GetClaimNameValidationQuery(ctx, *request.Spec.ClaimName, request.Meta.Scope,
			convutils.Deref(request.Meta.ProjectId, 0),
			convutils.Deref(request.Meta.DepartmentId, ""),
			convutils.Deref(request.Meta.ClusterId, ""))
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}
		claimNameValidationQuery = query
	}

	result, actionError := dao.CreatePvcAsset(ctx, request, nameValidationQuery, claimNameValidationQuery, dvId)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)

	err := addCreationEvent(ctx, result)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusAccepted, result), nil
}

func (s *PVCApiService) DeletePvcAssetById(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id)

	pvcAsset, getError := dao.GetPvcAssetById(ctx, id, dao.DatasourcesOptions{
		IncludeDeleted: true,
	})
	if getError != nil {
		return NewErrorResponse(getError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfMeta(&pvcAsset.Meta), string(assetKind), authClient.ACTION_DELETE, convutils.Deref(pvcAsset.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	validationError := validateNotBeingUsedByTemplate(ctx, id, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	if validationError = validatePvcNotBeingUsedByDatavolumes(ctx, pvcAsset); validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	if convutils.BoolIsTrue(pvcAsset.Spec.DataSharing) {
		err := deleteDatavolumeOfSharedPvc(ctx, pvcAsset.Meta.Id)
		if err != nil {
			return NewErrorResponse(err), nil
		}
	}

	//
	//   non-existing pvcs are softly deleted, in order to allow the agent to sync with the deletion
	//
	var actionError *api.Error

	if convutils.BoolIsTrue(pvcAsset.Spec.ExistingPvc) || !agentSyncSupported(ctx, pvcAsset.Meta.ClusterId) {
		actionError = dao.DeletePvcAsset(ctx, &pvcAsset.Meta, dao.HardDelete)
	} else {
		actionError = dao.DeletePvcAsset(ctx, &pvcAsset.Meta, dao.SoftDelete)
	}

	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	err := addDeletionEvent(ctx, pvcAsset)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(pvcAsset.Meta.Name).WithClusterId(safeStrPtr(pvcAsset.Meta.ClusterId))
	return NewDeletionResponse(), nil
}

func (s *PVCApiService) GetPvcAssetById(ctx context.Context, id string,
	usageInfo bool, complianceProject int32, complianceWorkloadType string, statusInfo bool, complianceReplicaType string, dvUsageInfo bool,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC

	options := dao.DatasourcesOptions{
		IncludeDeleted: statusInfo, // if we include status info we can include deleted assets and show their deletion status
	}

	asset, actionError := dao.GetPvcAssetById(ctx, id, options)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	permittedScopes, allowedProjectIds, authError := query_types.ApplyRbacRulesGetById(ctx, asset.Meta)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	if statusInfo || usageInfo || complianceProject != 0 || complianceWorkloadType != "" || dvUsageInfo {
		patchingPointers := response_patching.LoadResponsePatchingPtrs(asset)

		if usageInfo {
			patchErr := usage_info.PatchUsageInfo(ctx, assetKind, patchingPointers, allowedProjectIds)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}

		if dvUsageInfo {
			err := validatePvcNotBeingUsedByDatavolumes(ctx, asset)
			if err != nil {
				if err.Code == http.StatusPreconditionFailed {
					if asset.UsedBy == nil {
						asset.UsedBy = &api.AssetUsageInfo{}
					}
					asset.UsedBy.UsedByDataVolumes = convutils.Ref(true)
				} else {
					return NewErrorResponse(err), nil
				}
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
				patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_PVC)
				if patchErr != nil {
					return NewErrorResponse(patchErr), nil
				}
			}
		}
	}

	return api.Response(http.StatusOK, asset), nil
}

func (s *PVCApiService) ListPvcAssets(ctx context.Context, nameParam string, scopeParam string, projectIdParam int32,
	departmentId string, clusterId string, usageInfo bool, complianceProject int32, complianceWorkloadType string, statusInfo bool, assetIds string, complianceReplicaType string,
) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC

	queryParams, queryError := query_types.NewQueryParams(assetIds, nameParam, scopeParam, false, projectIdParam, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	options := dao.DatasourcesOptions{
		IncludeDeleted: statusInfo, // if we include status info we can include deleted assets and show their deletion status
	}

	result, actionError := dao.ListPvcAssets(ctx, queryParams, options)
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
			patchErr := status_info.PatchStatusInfo(ctx, assetClusterStatus, patchingPointers, api.ASSETKIND_PVC)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, result), nil
}

func (s *PVCApiService) UpdatePvcAssetById(ctx context.Context, id string, request api.PvcUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id).WithEntityName(request.Meta.Name)

	meta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(meta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(meta), string(assetKind), authClient.ACTION_UPDATE, convutils.Deref(meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	/* At this point we are supporting only container path so nothing to validate in the spec
	   validationErr := validatePvcSpec(ctx, &request.Meta, &request.Spec, true)
	   if validationErr != nil {
	   	return NewErrorResponse(validationErr), nil
	   }
	*/

	//
	//   validate the requested name, and if it is ok - update the meta with the requested name
	//
	nameValidationQuery, err := scope.GetNameValidationQuery(ctx, request.Meta.Name, meta.Scope, meta.ProjectId, meta.DepartmentId, meta.ClusterId, meta.Private, api.ASSETKIND_PVC, &id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	//
	//   load the spec from DB and update it with the new values
	//
	spec, dbErr := dao.GetPvcAssetSpecById(ctx, id)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	spec.Path = request.Spec.Path

	if request.Spec.ClaimInfo != nil {
		if spec.ClaimInfo == nil {
			spec.ClaimInfo = &api.ClaimInfo{}
		}

		spec.ClaimInfo.AddedAttrValues = request.Spec.ClaimInfo.AddedAttrValues
	}

	//
	//   now that we have the updated spec, need to check it and apply any storage class configuration thing
	//
	validateErr := validatePvcSpec(ctx, meta, &spec.PvcAssetSpec, true)
	if validateErr != nil {
		return NewErrorResponse(validateErr), nil
	}

	//
	//   special handling with PVC change of size, it requires specific validations (for example
	//   to check if it is bounded, etc)
	//
	validateErr = validateChangeOfPvcSize(ctx, id, meta, spec, &request)
	if validateErr != nil {
		return NewErrorResponse(validateErr), nil
	}

	result, actionError := dao.UpdatePvcAsset(ctx, id, request.Meta, spec, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusAccepted, result), nil
}

// GetPvcHistory - Get the PVC history.
func (s *PVCApiService) GetPvcHistory(ctx context.Context, assetId string, offset int32, limit int32) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_PVC
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(assetId)

	meta, getError := dao.GetAssetMetaById(ctx, assetId)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(meta.ClusterId))

	validationError := validateMinimumClusterVersionForPVCEvents(ctx, safeStrPtr(meta.ClusterId))
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(meta), string(assetKind), authClient.ACTION_READ, convutils.Deref(meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	events, actionError := dao.GetEventsByAssetId(ctx, assetId, offset, limit)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	apiHistory := []api.HistoryRecord{}
	for _, event := range events {
		apiHistory = append(apiHistory, eventModelToApi(event, meta))
	}

	response := api.PvcHistoryResponse{
		Records: apiHistory,
	}
	recordsCount := int32(len(apiHistory))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func validatePvcSpec(ctx context.Context, meta *dbmodel.AssetMeta, pvc *api.PvcAssetSpec, isUpdate bool) *api.Error {
	if pvc == nil {
		return nil
	}

	if convutils.BoolIsTrue(pvc.ExistingPvc) {
		if convutils.BoolIsTrue(pvc.DataSharing) {
			return api.NewValidationError("dataSharing cannot be enabled when existingPvc is true")
		}
		if pvc.ClaimInfo != nil {
			return api.NewValidationError("claimInfo should not be provided for existingPvc")
		}
		err := validatePvcAssetVsClusterResource(ctx, meta, pvc.ClaimName)
		if err != nil {
			return err
		}
		return nil
	}

	// The code from here can assume that the PVC asset is a new PVC (not base on existing claim)

	if pvc.ClaimInfo == nil {
		return api.NewValidationError("claimInfo is mandatory when existingPvc = false")
	}

	if pvc.ClaimName == nil {
		return api.NewValidationError("claimName is mandatory when existingPvc = false")
	}

	// skip on update: the asset's own PVC already exists in the cluster, so the uniqueness check always finds it and returns a false 409
	if !isUpdate {
		claimNameErr := validateClaimNameVsClusterResource(ctx, meta, pvc.ClaimName)
		if claimNameErr != nil {
			return claimNameErr
		}
	}

	if pvc.ClaimInfo.StorageClass == nil {
		return flat_workloads.MakeSureNoAddedAttrs(pvc.ClaimInfo)
	}

	// storage class is per cluster id, so we must know the cluster id
	if meta.ClusterId == nil {
		return flat_workloads.MakeSureNoAddedAttrs(pvc.ClaimInfo)
	}

	if pvc.DataSharing != nil && *pvc.DataSharing {
		if meta.Scope == api.SCOPE_PROJECT {
			return api.NewError(http.StatusBadRequest, "dataSharing is not supported for project scoped PVC", nil)
		}

		err := validateMinimumClusterVersionForPVCDataSharing(ctx, *meta.ClusterId)
		if err != nil {
			return err
		}
	}

	config, err := k8sObjectsTracker.Get().GetStorageClassConfigInt(ctx, *meta.ClusterId, *pvc.ClaimInfo.StorageClass)
	if err != nil {
		// if there is no configuration, we ignore the configuration
		if err.Code == http.StatusNotFound {
			return flat_workloads.MakeSureNoAddedAttrs(pvc.ClaimInfo)
		}

		return api.NewError(int(err.Code), err.Message, nil)
	}

	//
	//   checking the permission of creating the asset depends on the type of auto deletions. auto deleted assets
	//   are volumes, and so they should be validated VS the volume creation permission. other assets are validated
	//   VS the asset creation permission
	//
	if config.Permissions.Get() != nil {
		var whatToCheck k8sObjectsApi.NullableBool
		var whatToReport string

		if convutils.BoolIsTrue(meta.AutoDelete) {
			if convutils.BoolIsTrue(pvc.Ephemeral) {
				whatToCheck = config.Permissions.Get().AllowedForEphemeralVolumes
				whatToReport = "ephemeral volume creation"
			} else {
				whatToCheck = config.Permissions.Get().AllowedForPersistentVolumes
				whatToReport = "persistent volume creation"
			}
		} else {
			whatToCheck = config.Permissions.Get().AllowedForAssets
			whatToReport = "creating PVCs"
		}

		if convutils.BoolIsFalse(whatToCheck.Get()) {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("%s is excluded from %s", StorageClassNameForMessage(pvc), whatToReport), nil)
		}

		if convutils.BoolIsTrue(pvc.DataSharing) && convutils.BoolIsFalse(config.Permissions.Get().AllowDataSharing.Get()) {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("%s is not allowing data sharing", StorageClassNameForMessage(pvc)), nil)
		}
	}

	configErr := flat_workloads.ValidatePvcSpecVsScCustomization(meta.AutoDelete, pvc, config)
	if configErr != nil {
		return configErr
	}

	return nil
}

func validateChangeOfPvcSize(ctx context.Context, id string, meta *dbmodel.AssetMeta, spec *dbmodel.PVCAssetRec, request *api.PvcUpdateRequest) *api.Error {
	// currently we validate only the size so no size no validation
	if request == nil || request.Spec == nil || request.Spec.ClaimInfo == nil || request.Spec.ClaimInfo.Size == nil {
		return nil
	}

	if convutils.BoolIsTrue(spec.ExistingPvc) {
		return api.NewValidationError("claimInfo.size is permitted only for PVC assets with existingPvc=false")
	}

	//
	//  validation #1: requested updated size must be no less than the originally requested size
	//
	fromSizeAsStr := convutils.Deref(spec.ClaimInfo.Size, "")
	toSizeAsStr := *request.Spec.ClaimInfo.Size

	fromSize, convErr := resource.ParseQuantity(fromSizeAsStr)
	if convErr != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to parse PVC existing size %s", fromSizeAsStr), convErr)
	}

	toSize, convErr := resource.ParseQuantity(toSizeAsStr)
	if convErr != nil {
		return api.NewValidationError(fmt.Sprintf("malformed PVC size %s: %v", toSizeAsStr, convErr))
	}

	if toSize.Value() < fromSize.Value() {
		return api.NewValidationError(fmt.Sprintf("requested size (%s) must be no less than the originakl requested size (%s) of the PVC", toSizeAsStr, fromSizeAsStr))
	}

	//
	//  validation #2: validate PVC is Bound
	//
	if meta.ClusterId == nil {
		return nil
	}

	var scopeId string
	switch meta.Scope {
	case api.SCOPE_CLUSTER:
		scopeId = *meta.ClusterId
	case api.SCOPE_DEPARTMENT:
		scopeId = convutils.Deref(meta.DepartmentId, "")
	case api.SCOPE_PROJECT:
		scopeId = fmt.Sprintf("%d", convutils.Deref(meta.ProjectId, 0))
	}

	pvcs, trackerErr := k8sObjectsTracker.Get().ListPvcsByScopeAndName(ctx, string(meta.Scope), scopeId, convutils.Deref(spec.ClaimName, ""))
	if trackerErr != nil {
		if trackerErr.Code == http.StatusNotFound {
			return api.NewError(http.StatusNotFound, "PVC is not available in the cluster yet", nil)
		}
		return api.NewError(int(trackerErr.Code), trackerErr.Message, nil)
	}

	// in practice, we'll either have 0 or 1 results, but since list returns an array we run over the array
	for _, pvc := range pvcs.Pvcs {
		if pvc.Phase.Get() == nil {
			continue
		}

		if *pvc.Phase.Get() != k8sObjectsApi.PVCPHASEOPTIONAL_BOUND {
			return api.NewValidationError(fmt.Sprintf("resize is only allowed for Bound PVCs, current PVC state is %s", *pvc.Phase.Get()))
		}

		if pvc.RequestedCapacity.Get() == nil {
			continue
		}

		reqCapacityStr := *pvc.RequestedCapacity.Get()

		reqCapacity, convErr := resource.ParseQuantity(reqCapacityStr)
		if convErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("invalid requested capqcity size %s", reqCapacityStr)
		} else {
			if toSize.Value() < reqCapacity.Value() {
				return api.NewValidationError(fmt.Sprintf("requested size (%s) must be no less than the effective requested capacity (%s) of the PVC", reqCapacityStr, toSizeAsStr))
			}
		}
	}

	//  update the spec with the new size
	spec.ClaimInfo.Size = request.Spec.ClaimInfo.Size

	return nil
}

// validateEphemeralAutoDelete validate if ephemeral is set to true, then autoDelete should be set to true as well
func validateEphemeralAutoDelete(pvc *api.PvcAssetSpec, meta *api.AssetCreationRequest) *api.Error {
	if pvc == nil {
		return nil
	}

	if convutils.Deref(pvc.Ephemeral, false) {
		if meta == nil || !convutils.Deref(meta.AutoDelete, false) {
			return api.NewValidationError("autoDelete is mandatory when ephemeral is true")
		}
	}

	return nil
}

func validatePvcAssetVsClusterResource(ctx context.Context, meta *dbmodel.AssetMeta, resourceName *string) *api.Error {
	if resourceName == nil {
		return nil
	}

	scopeInfo := ScopeInfoOfDbMeta(meta)

	resp, err := k8sObjectsTracker.Get().ListPvcsByScopeAndName(ctx, string(scopeInfo.Scope), scopeInfo.ScopeId, *resourceName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to inquire PVC status: %v", err.Message)
		return api.NewError(http.StatusServiceUnavailable, "failed to inquire the status of the PVC in the cluster", errors.New(err.Message))
	}

	if len(resp.Pvcs) == 0 {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("PVC resource '%s' of %s scope could not be found in the cluster", *resourceName, scopeInfo.Scope), nil)
	}

	if convutils.HasValue(resp.Pvcs[0].Origin) {
		if *resp.Pvcs[0].Origin.Get() != k8sObjectsApi.ORIGIN_EXISTING {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("cannot create an existingPvc=true asset on resource '%s' because it is linked to an existingPvc=false asset", *resourceName), nil)
		}
	}

	return nil
}

// validateClaimNameVsClusterResource checks if a PVC with the given claim name already exists in the cluster.
// This is used to prevent creating a non-existing PVC asset when a PVC with that name already exists in the cluster.
func validateClaimNameVsClusterResource(ctx context.Context, meta *dbmodel.AssetMeta, claimName *string) *api.Error {
	if claimName == nil {
		return nil
	}

	scopeInfo := ScopeInfoOfDbMeta(meta)

	resp, err := k8sObjectsTracker.Get().ListPvcsByScopeAndName(ctx, string(scopeInfo.Scope), scopeInfo.ScopeId, *claimName)
	if err != nil {
		// Graceful check, if there was an issue with comm to the k8s tracker, log it, but don't fail the asset creation
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to inquire PVC status from k8s tracker: %v", err.Message)
		return nil
	}

	if len(resp.Pvcs) > 0 {
		return api.NewError(http.StatusConflict, fmt.Sprintf("PVC resource '%s' already exists in the %s scope in the cluster", *claimName, scopeInfo.Scope), nil)
	}

	return nil
}

func validateMinimumClusterVersionForPVCDataSharing(ctx context.Context, clusterId string) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfPVCDataSharing)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "PVC Datasharing is not supported in this cluster version", nil)
	}

	return nil
}

func validateMinimumClusterVersionForPVCEvents(ctx context.Context, clusterId string) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfPVCEvents)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "PVC Events are not supported in this cluster version", nil)
	}

	return nil
}

func createDatavolumeForSharedPvc(ctx context.Context, request api.PvcCreationRequest) (*string, *api.Error) {
	if request.Spec.ClaimName == nil {
		return nil, api.NewError(http.StatusBadRequest, "missing mandatory claim name for PVC data sharing", nil)
	}

	dvCreationRequest := openapi.InternalDatavolumeCreationFields{
		Name:          request.Meta.Name,
		Scope:         openapi.ScopeType(request.Meta.Scope),
		ClusterId:     request.Meta.ClusterId,
		DepartmentId:  request.Meta.DepartmentId,
		OriginPvcName: *request.Spec.ClaimName,
	}

	dv, resp, err := datavolumes_service.Get().CreateInternalDatavolume(ctx, dvCreationRequest)
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		if resp != nil {
			if resp.StatusCode == http.StatusConflict {
				return nil, api.NewError(resp.StatusCode, fmt.Sprintf("PVC with claim name '%s' was already shared in this scope", *request.Spec.ClaimName), nil)
			}
			statusCode = (*resp).StatusCode
		}
		return nil, api.NewError(statusCode, err.Error(), err)
	}

	return &dv.Id, nil
}

func deleteDatavolumeOfSharedPvc(ctx context.Context, pvcAssetId string) *api.Error {
	dvId, err := dao.GetPvcDatavolumeId(ctx, pvcAssetId)
	if err != nil {
		return err
	}

	_, resp, apiErr := datavolumes_service.Get().DeleteDatavolumeById(ctx, *dvId)
	if apiErr != nil {
		statusCode := http.StatusServiceUnavailable
		if resp != nil {
			statusCode = resp.StatusCode
		}

		return api.NewError(statusCode, apiErr.Error(), apiErr)
	}

	return nil
}

func eventModelToApi(event dbmodel.Event, meta *dbmodel.AssetMeta) api.HistoryRecord {
	historyRecord := api.HistoryRecord{
		Meta: api.HistoryRecordMeta{
			CreationTimestamp: event.CreatedAt,
		},
	}

	if event.Source == dbmodel.EVENT_SOURCE_ASSETS_SERVICE {
		historyRecord.Meta.Type = api.HISTORYRECORDTYPE_PHASE_UPDATE
		historyRecord.Spec = api.HistoryRecordSpec{
			PhaseUpdate: &api.AssetPhaseUpdate{
				Phase:        event.Reason,
				PhaseMessage: event.Message,
			},
		}
	} else {
		historyRecord.Meta.Type = api.HISTORYRECORDTYPE_EVENT
		historyRecord.Spec = api.HistoryRecordSpec{
			Event: &api.Event{
				CreatedAt: event.CreatedAt,
				Id:        event.ID.String(),
				Type:      event.Type,
				ClusterId: event.ClusterID.String(),
				Message:   event.Message,
				Reason:    event.Reason,
				Source:    event.Source,
			},
		}

		if event.ObjectKind != "" || event.ObjectName != "" || event.ObjectNamespace != "" || event.ObjectUid != "" {
			historyRecord.Spec.Event.InvolvedObject = api.InvolvedObject{
				Kind:      event.ObjectKind,
				Name:      event.ObjectName,
				Namespace: event.ObjectNamespace,
				Uid:       event.ObjectUid,
			}
		}
	}

	return historyRecord
}

func addCreationEvent(ctx context.Context, pvcAsset *api.PvcAsset) *api.Error {
	scopeId := ""
	switch pvcAsset.Meta.Scope {
	case api.SCOPE_TENANT:
		scopeId = fmt.Sprintf("%d", convutils.Deref(pvcAsset.Meta.TenantId, 0))
	case api.SCOPE_CLUSTER:
		scopeId = *pvcAsset.Meta.ClusterId
	case api.SCOPE_DEPARTMENT:
		scopeId = *pvcAsset.Meta.DepartmentId
	case api.SCOPE_PROJECT:
		scopeId = fmt.Sprintf("%d", convutils.Deref(pvcAsset.Meta.ProjectId, 0))
	}

	eventModel := dbmodel.Event{
		AssetId:   uuid.MustParse(pvcAsset.Meta.Id),
		CreatedAt: time.Now(),
		ID:        uuid.New(),
		TenantID:  uint(auth.TenantIdOf(ctx)),
		ScopeType: string(pvcAsset.Meta.Scope),
		ScopeId:   scopeId,
		AssetKind: string(api.ASSETKIND_PVC),
		Source:    dbmodel.EVENT_SOURCE_ASSETS_SERVICE,
		Reason:    "Creating",
		Type:      dbmodel.EVENT_TYPE_NORMAL,
		Message:   fmt.Sprintf("Creating PVC asset %s", pvcAsset.Meta.Name),
	}

	if pvcAsset.Meta.ClusterId != nil {
		eventModel.ClusterID = uuid.MustParse(*pvcAsset.Meta.ClusterId)
	}

	return dao.UpsertEvents(
		ctx,
		[]*dbmodel.Event{
			&eventModel,
		},
	)
}

func addDeletionEvent(ctx context.Context, pvcAsset *api.PvcAsset) *api.Error {
	scopeInfo := scopeInfo.ScopeInfoOf(scopeInfo.Scope(pvcAsset.Meta.Scope), pvcAsset.Meta.ProjectId, pvcAsset.Meta.DepartmentId, pvcAsset.Meta.ClusterId, pvcAsset.Meta.TenantId)
	scopeId := scopeInfo.ScopeId

	eventModel := dbmodel.Event{
		AssetId:   uuid.MustParse(pvcAsset.Meta.Id),
		CreatedAt: time.Now(),
		ID:        uuid.New(),
		TenantID:  uint(auth.TenantIdOf(ctx)),
		ScopeType: string(pvcAsset.Meta.Scope),
		ScopeId:   scopeId,
		AssetKind: string(api.ASSETKIND_PVC),
		Source:    dbmodel.EVENT_SOURCE_ASSETS_SERVICE,
		Reason:    "Deleting",
		Type:      dbmodel.EVENT_TYPE_NORMAL,
		Message:   fmt.Sprintf("Deleting PVC asset %s", pvcAsset.Meta.Name),
	}

	return dao.UpsertEvents(
		ctx,
		[]*dbmodel.Event{
			&eventModel,
		},
	)
}
