/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/google/uuid"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	datavolumes_cache "github.com/run-ai/runai/control-plane/datavolumes-service/pkg/cache"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/services_utils"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
)

type DatavolumesApiService struct {
	orgHierarchyManager *datavolumes_cache.OrgHierarchyManager
	authorizationClient authorization_client.IAuthorizationClient
}

type UsedPVCsResponse struct {
	UsedPVCs []string `json:"usedPVCs"`
}

// NewDatavolumesApiService creates a default api service
func NewDatavolumesApiService() api.DatavolumesAPIServicer {
	return &DatavolumesApiService{
		orgHierarchyManager: datavolumes_cache.GetOrgHierarchyManager(),
		authorizationClient: authorization_client.NewAuthorizationClient(),
	}
}

func (s *DatavolumesApiService) CreateDatavolume(ctx context.Context, datavolumeCreationFields api.DatavolumeCreationFields) (api.ImplResponse, error) {
	tenantId, subjectId, apiErr := services_utils.ParseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume create request")

	apiErr = ValidateCreationRequestDefaults(&datavolumeCreationFields)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	ancestors, _, err := s.orgHierarchyManager.GetAncestors(ctx, datavolumeCreationFields.ProjectId, api.SCOPETYPE_PROJECT)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", err)), nil
	}
	clusterId, err := uuid.Parse(*ancestors.ClusterUuid)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Failed to parse clusterId <%s> for project %s",
			*ancestors.ClusterUuid, datavolumeCreationFields.ProjectId), err)), nil
	}

	apiErr = validateMinimumClusterVersionForDatavolume(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue CreateDatavolume with nil permitted scopes. error: %s", err.Error())
	} else {
		if !authorization_utils.IsActionOnResourcePermittedByScopes(ancestors, permittedScopes) {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
		}
	}

	projectName, _, apiErr := s.orgHierarchyManager.GetScopeName(ctx, datavolumeCreationFields.ProjectId, api.SCOPETYPE_PROJECT, *ancestors.ClusterUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	datavolume, apiErr := dao.DbCreateDatavolume(ctx, DatavolumeCreateRequestToModel(
		datavolumeCreationFields, tenantId, subjectId, clusterId, projectName, *ancestors.DepartmentId))
	if apiErr != nil {
		if apiErr.Code == http.StatusConflict {
			return parseConflictOnCreateError(apiErr), nil
		}
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully created datavolume <%s> <%s/%s>",
		datavolume.ID, datavolume.Name, datavolume.Description)
	return api.Response(http.StatusCreated, DatavolumeModelToApi(*datavolume)), nil
}

func (s *DatavolumesApiService) DeleteDatavolume(ctx context.Context, datavolumeId string) (api.ImplResponse, error) {
	tenantId, subjectId, apiErr := services_utils.ParseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> delete request", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	datavolumeMeta, apiErr := dao.DbGetDatavolumeById(ctx, tenantId, id, true)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if len(datavolumeMeta.SharedScopes) > 0 {
		ancestors := datavolumeMetaToAncestors(datavolumeMeta)
		hasUpdateSharedScopesPermission, innerErr := s.doesHaveUpdateSharedScopesPermission(ctx, ancestors)
		if innerErr != nil {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", innerErr)), nil
		}
		if !hasUpdateSharedScopesPermission {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Not allowed to delete datavolume with shared scopes", nil)), nil
		}
	}

	if datavolumeMeta.Phase == api.DATAVOLUMEPHASE_CREATING {
		// we don't want the DV to be stuck in "Creating" - we need to preform the actual Deletion in this case.
		// even if the cluster recognized it and will send an update status request soon,
		// the next time the agent picks it up it should delete the DV from the cluster as well.
		apiErr = dao.DbDeleteDatavolume(ctx, tenantId, id)
		if apiErr != nil {
			if apiErr.Code != http.StatusNotFound {
				return NewErrorResponse(apiErr), nil
			}
			// StatusNotFound normally shouldn't happen - since we just got the datavolumeMeta,
			// so this could only happen if there was a race condition and the DV's phase was updated in the meantime.
			// in this case we want to continue and update the phase to deleting.
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted datavolume <%s>", datavolumeId)
			return NoContentResponse(), nil
		}
	}

	apiErr = dao.DbUpdatePhaseToDeleting(ctx, tenantId, id, subjectId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully marked datavolume <%s> as Deleting", datavolumeId)
	return api.Response(http.StatusAccepted, api.HttpResponse{Code: http.StatusAccepted, Message: "Datavolume is marked as 'Deleting'; pending deletion from the cluster"}), nil
}

func (s *DatavolumesApiService) GetDatavolume(ctx context.Context, datavolumeId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> get request", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	datavolume, apiErr := dao.DbGetDatavolumeById(ctx, tenantId, id, true)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue GetDatavolume with nil permitted scopes. error: %s", err.Error())
		permittedScopes = nil
	}

	permittedSharedScopes := filterOutNonPermittedSharedScopes(tenantId, datavolume.ClusterId, datavolume.SharedScopes, permittedScopes)
	datavolume.SharedScopes = permittedSharedScopes

	return api.Response(http.StatusOK, DatavolumeModelToApi(*datavolume)), nil
}

func (s *DatavolumesApiService) GetDatavolumes(
	ctx context.Context,
	requestType api.DatavolumeRequestType,
	usableInProjectId string,
	offset int32,
	limit int32,
	sortBy string,
	sortOrder string,
	filterBy []string,
	search string,
	includeInternal bool,
) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolumes get request - requestType: %v, usableInProjectId: %s, offset: %d, limit: %d, sortBy: %s, sortOrder: %s, filterBy: %v",
		requestType, usableInProjectId, offset, limit, sortBy, sortOrder, filterBy)

	apiErr = validateGetVolumesRequest(requestType, usableInProjectId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	parser := queryapi.NewParser(models.DatavolumeView{}.SortAndFilterFieldsToColumn())
	filters, err := parser.Filters(filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate filterBy param, error: %s", err.Error())
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}

	sortColumn, err := parser.GetMapping(sortBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate sortBy param, error: %s", err.Error())
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}

	limit = setDefaultLimit(limit)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue GetDatavolumes with nil permitted scopes. error: %s", err.Error())
		permittedScopes = nil
	}

	var usableInProjectIdAncestors *authorization_utils.Ancestors = nil
	if requestType == api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT {
		usableInProjectIdAncestors, _, err = s.orgHierarchyManager.GetAncestors(ctx, usableInProjectId, api.SCOPETYPE_PROJECT)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", err)), nil
		}

		// validate the user has permissions to list datavolumes shared with the requested project
		if !authorization_utils.IsActionOnResourcePermittedByScopes(usableInProjectIdAncestors, permittedScopes) {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
		}
	}

	datavolumes, apiErr := dao.DbGetDatavolumes(ctx, tenantId,
		permittedScopes,
		usableInProjectIdAncestors,
		offset, limit, sortColumn, sortOrder, filters, search, includeInternal)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	response := api.GetDatavolumes200Response{
		Datavolumes: convutils.ConvertSliceTypes(datavolumes, DatavolumeViewModelToApiNoSharedScopes),
	}
	recordsCount := int32(len(datavolumes))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

// CountDatavolumes - Count data volumes.
func (s *DatavolumesApiService) CountDatavolumes(ctx context.Context, requestType api.DatavolumeRequestType, usableInProjectId string, filterBy []string, search string) (api.ImplResponse, error) {

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolumes count request - requestType: %v, usableInProjectId: %s, filterBy: %v",
		requestType, usableInProjectId, filterBy)

	apiErr = validateGetVolumesRequest(requestType, usableInProjectId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	parser := queryapi.NewParser(models.DatavolumeView{}.SortAndFilterFieldsToColumn())
	filters, err := parser.Filters(filterBy)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to validate filterBy param, error: %s", err.Error())
		return ErrorToApiError(err, http.StatusBadRequest), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue GetDatavolumes with nil permitted scopes. error: %s", err.Error())
		permittedScopes = nil
	}

	var usableInProjectIdAncestors *authorization_utils.Ancestors = nil
	if requestType == api.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT {
		usableInProjectIdAncestors, _, err = s.orgHierarchyManager.GetAncestors(ctx, usableInProjectId, api.SCOPETYPE_PROJECT)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", err)), nil
		}

		// validate the user has permissions to list datavolumes shared with the requested project
		if !authorization_utils.IsActionOnResourcePermittedByScopes(usableInProjectIdAncestors, permittedScopes) {
			return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
		}
	}

	count, dbErr := dao.DbCountDatavolumes(ctx, tenantId, permittedScopes, usableInProjectIdAncestors, filters, search)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountDatavolumes200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DatavolumesApiService) GetDatavolumeSharedScopes(ctx context.Context, datavolumeId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> get sharedscopes request", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	datavolume, apiErr := dao.DbGetDatavolumeById(ctx, tenantId, id, true)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue GetDatavolumeSharedScopes with nil permitted scopes. error: %s", err.Error())
		permittedScopes = nil
	}

	permittedSharedScopes := filterOutNonPermittedSharedScopes(tenantId, datavolume.ClusterId, datavolume.SharedScopes, permittedScopes)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully got datavolume <%s> sharedscopes with len: %d", datavolumeId, len(permittedSharedScopes))
	return api.Response(http.StatusOK, sharedScopesModelToApi(permittedSharedScopes)), nil
}

func (s *DatavolumesApiService) GetDatavolumesInternal(ctx context.Context, clusterId string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolumes internal get request - clusterId: %s, offset: %d, limit: %d", clusterId, offset, limit)

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("clusterId must be a valid uuid: <%s>", clusterId), err)), nil
	}

	limit = setDefaultLimit(limit)

	datavolumes, apiErr := dao.DbGetDatavolumesInClusterWithPartialStatus(ctx, tenantId, clusterUuid, offset, limit)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	response := api.GetDatavolumesInternal200Response{
		Datavolumes: convutils.ConvertSliceTypes(datavolumes, DatavolumeModelToApiInternal),
	}
	recordsCount := int32(len(datavolumes))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DatavolumesApiService) PatchDatavolumeSharedScopes(ctx context.Context, datavolumeId string, sharedScopesPatchRequest api.SharedScopesPatchRequest) (api.ImplResponse, error) {
	tenantId, subjectId, apiErr := services_utils.ParseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> patch shared scopes", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	datavolume, apiErr := dao.DbGetDatavolumeById(ctx, tenantId, id, true)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to parse permitted scopes; will continue PatchDatavolumeSharedScopes with nil permitted scopes. error: %s", err.Error())
		permittedScopes = nil
	}

	existingSharedScopes := datavolume.SharedScopes
	toCreate, toKeep, toDelete, apiErr := s.convertAndValidateSharedScopesUpsertRequest(ctx, tenantId, subjectId, datavolume,
		permittedScopes, sharedScopesPatchRequest, existingSharedScopes)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	created, apiErr := dao.DbUpsertDatavolumeSharedScopes(ctx, tenantId, id, toCreate, toDelete)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully upserted datavolume <%s> shared scopes", datavolumeId)
	result := append(toKeep, created...)
	result = filterOutNonPermittedSharedScopes(tenantId, datavolume.ClusterId, result, permittedScopes)
	return api.Response(http.StatusOK, sharedScopesModelToApi(result)), nil
}

func (s *DatavolumesApiService) PatchDatavolume(ctx context.Context, datavolumeId string, datavolumePatchFields api.DatavolumePatchFields) (api.ImplResponse, error) {
	tenantId, subjectId, apiErr := services_utils.ParseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> patch", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	datavolume, apiErr := dao.DbPatchDatavolume(ctx, tenantId, id, subjectId, datavolumePatchFields.Description)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully patched datavolume <%s> with description: <%s>", datavolumeId, datavolumePatchFields.Description)
	return api.Response(http.StatusOK, DatavolumeModelToApiNoSharedScopes(*datavolume)), nil
}

func (s *DatavolumesApiService) SyncDatavolumes(ctx context.Context, datavolumesSyncRequest api.DatavolumesSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolumes sync request for cluster <%s>; from: %s, to: %s",
		datavolumesSyncRequest.ClusterId, datavolumesSyncRequest.From, datavolumesSyncRequest.To)

	from, to, apiErr := parseFromToUuids(datavolumesSyncRequest.From, datavolumesSyncRequest.To)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	datavolumeStatuses := []models.Datavolume{}
	for _, datavolume := range datavolumesSyncRequest.Datavolumes {
		datavolumeWithStatus, err := datavolumeStatusForSyncToModelWithStatus(datavolume)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Failed to parse datavolume <%s> status", datavolume.DatavolumeId), err)), nil
		}
		datavolumeStatuses = append(datavolumeStatuses, datavolumeWithStatus)
	}

	apiErr = dao.DbSyncDatavolumeStatuses(ctx, tenantId, datavolumesSyncRequest.ClusterId, datavolumeStatuses, from, to)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully synced %d datavolumes for cluster <%s>; from: %s, to: %s",
		len(datavolumeStatuses), datavolumesSyncRequest.ClusterId, datavolumesSyncRequest.From, datavolumesSyncRequest.To)
	return NoContentResponse(), nil
}

func (s *DatavolumesApiService) UpdateDatavolumeStatus(ctx context.Context, datavolumeId string, datavolumeStatusInternal api.DatavolumeStatusInternal) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting datavolume <%s> update status request", datavolumeId)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)), nil
	}

	if datavolumeStatusInternal.IsDeleted != nil && *datavolumeStatusInternal.IsDeleted {
		apiErr = dao.DbDeleteDatavolume(ctx, tenantId, id)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted datavolume <%s> after deletion in cluster", datavolumeId)
		return NoContentResponse(), nil
	}

	apiErr = dao.DbUpdateStatus(ctx, tenantId, id, DatavolumeStatusApiToModel(datavolumeStatusInternal))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully updated status of datavolume <%s> to: %s/%s/%s/%s",
		datavolumeId, datavolumeStatusInternal.Phase, datavolumeStatusInternal.PhaseMessage, datavolumeStatusInternal.DatavolumePvcName, datavolumeStatusInternal.DatavolumePvName)
	return NoContentResponse(), nil
}

// DatavolumeNameAvailability - Data volumes name availability.
func (s *DatavolumesApiService) DatavolumeNameAvailability(ctx context.Context, dataName string, clusterId string) (api.ImplResponse, error) {
	isOk, err := dao.DatavolumeNameAvailability(ctx, dataName, clusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	if !isOk {
		return api.Response(http.StatusConflict, api.NewError(http.StatusConflict, "the given data-volume name is not available", nil)), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

// GetUsedPvcsFromDatavolumes - Get list of PVC names used by DataVolumes in the specified project
func (s *DatavolumesApiService) GetUsedPvcsFromDatavolumes(ctx context.Context, projectId string) (api.ImplResponse, error) {
	if projectId == "" {
		return api.Response(http.StatusBadRequest, api.NewValidationError("Missing mandatory projectId parameter")), nil
	}

	pvcs, apiErr := dao.GetUsedPvcsInPermittedScopes(ctx, projectId)
	if apiErr != nil {
		return api.Response(http.StatusInternalServerError, api.NewValidationError(apiErr.Message)), nil
	}

	return api.Response(http.StatusOK, UsedPVCsResponse{UsedPVCs: pvcs}), nil
}

func (s *DatavolumesApiService) GetDatavolumeAncestors(ctx context.Context, datavolumeId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	ancestors, apiErr := s.getDatavolumeAncestors(ctx, tenantId, datavolumeId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, *ancestors), nil
}

// InternalCreateDatavolume - Create a datavolume
// This function does not perform authorization checks because it is intended for internal use only.
// Any requests reaching this internal API are assumed to have already passed authorization
// in the originating service.
// This is currently a side effect of creating a PVC.
// The user who created the PVC may not necessarily have permission to create a DV.
func (s *DatavolumesApiService) InternalCreateDatavolume(ctx context.Context, internalDatavolumeCreationFields api.InternalDatavolumeCreationFields) (api.ImplResponse, error) {
	if internalDatavolumeCreationFields.Scope == api.SCOPETYPE_PROJECT {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "project scope datavolume is not supported in internal API", nil)), nil
	}

	if internalDatavolumeCreationFields.OriginPvcName == "" {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "originPvcName is mandatory", nil)), nil
	}

	var clusterId uuid.UUID
	var departmentName string
	var clusterName string

	if internalDatavolumeCreationFields.Scope == api.SCOPETYPE_DEPARTMENT {
		if internalDatavolumeCreationFields.DepartmentId == "" {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, "departmentId is mandatory for department scoped datavolume", nil)), nil
		}

		ancestors, _, err := s.orgHierarchyManager.GetAncestors(ctx, internalDatavolumeCreationFields.DepartmentId, api.SCOPETYPE_DEPARTMENT)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, err.Error(), err)), nil
		}

		clusterId, err = uuid.Parse(*ancestors.ClusterUuid)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Failed to parse clusterId <%s> for department %s",
				*ancestors.ClusterUuid, internalDatavolumeCreationFields.DepartmentId), err)), nil
		}

		var apiErr *api.Error
		departmentName, _, apiErr = s.orgHierarchyManager.GetScopeName(ctx, internalDatavolumeCreationFields.DepartmentId, api.SCOPETYPE_DEPARTMENT, *ancestors.ClusterUuid)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}
	}

	if internalDatavolumeCreationFields.Scope == api.SCOPETYPE_CLUSTER {
		if internalDatavolumeCreationFields.ClusterId == "" {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, "clusterId is mandatory for cluster scoped datavolume", nil)), nil
		}

		var err error
		clusterId, err = uuid.Parse(internalDatavolumeCreationFields.ClusterId)
		if err != nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Failed to parse clusterId <%s>",
				internalDatavolumeCreationFields.ClusterId), err)), nil
		}

		var apiErr *api.Error
		clusterName, _, apiErr = s.orgHierarchyManager.GetScopeName(ctx, internalDatavolumeCreationFields.ClusterId, api.SCOPETYPE_CLUSTER, internalDatavolumeCreationFields.ClusterId)
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}
	}

	apiErr := validateMinimumClusterVersionForClusterAndDepartmentScopeDatavolume(ctx, clusterId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, _, apiErr := services_utils.ParseAuthData(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	datavolumeDbModel := internalDatavolumeCreateRequestToModel(internalDatavolumeCreationFields, tenantId, clusterId, departmentName)
	datavolume, apiErr := dao.DbCreateAndShareDatavolume(ctx, tenantId, datavolumeDbModel, clusterName)
	if apiErr != nil {
		if apiErr.Code == http.StatusConflict {
			return parseConflictOnCreateError(apiErr), nil
		}
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, datavolume), nil
}

func (s *DatavolumesApiService) getDatavolumeAncestors(ctx context.Context, tenantId uint, datavolumeId string) (*authorization_utils.Ancestors, *api.Error) {
	ancestors, err := authorization_utils.GetAncestorsCache(ctx, string(authorization_api.RESOURCETYPE_DATAVOLUMES), datavolumeId)
	if err == nil && ancestors != nil {
		return ancestors, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find datavolume <%s> ancestors in cache. err: %v, continuing", datavolumeId, err)

	id, err := uuid.Parse(datavolumeId)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("datavolumeId must be a valid uuid: <%s>", datavolumeId), err)
	}

	datavolumeMeta, apiErr := dao.DbGetDatavolumeMeta(ctx, tenantId, id)
	if apiErr != nil {
		return nil, apiErr
	}

	ancestors = datavolumeMetaToAncestors(datavolumeMeta)
	authorization_utils.SetAncestorsCache(ctx, string(authorization_api.RESOURCETYPE_DATAVOLUMES), datavolumeId, *ancestors)

	return ancestors, nil
}

func (s *DatavolumesApiService) doesHaveUpdateSharedScopesPermission(ctx context.Context, datavolumeAncestors *authorization_utils.Ancestors) (bool, error) {
	params := authorization_api.NewBackendEvaluationRequest()
	params.SetRequestAction(authorization_api.ACTION_UPDATE)
	params.SetResourceType(authorization_api.RESOURCETYPE_DATAVOLUMES)

	request := s.authorizationClient.EvaluatePermission(ctx)
	request = request.BackendEvaluationRequest(*params)
	permittedScopes, response, err := s.authorizationClient.EvaluatePermissionExecute(request)
	if err != nil {
		return false, err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return false, fmt.Errorf("failed to evaluate permission, status code: %d", response.StatusCode)
	}

	if permittedScopes == nil {
		return false, nil
	}

	return authorization_utils.IsActionOnResourcePermittedByScopes(datavolumeAncestors, permittedScopes), nil
}

func (s *DatavolumesApiService) convertAndValidateSharedScopesUpsertRequest(ctx context.Context, tenantId uint, subjectId string, datavolume *models.Datavolume,
	permittedScopes *authorization_api.PermittedScopes,
	sharedScopesPatchRequest api.SharedScopesPatchRequest, existingSharedScopes []models.SharedScope,
) (toCreate []models.SharedScope, toKeep []models.SharedScope, toDelete []models.SharedScope, err *api.Error) {
	requestedToAdd := sharedScopesBasesToMap(sharedScopesPatchRequest.Add)
	if len(requestedToAdd) != len(sharedScopesPatchRequest.Add) {
		return nil, nil, nil, api.NewError(http.StatusConflict, "Duplicate shared scopes in requested Add", nil)
	}

	requestedToRemove := sharedScopesBasesToMap(sharedScopesPatchRequest.Remove)
	if len(requestedToRemove) != len(sharedScopesPatchRequest.Remove) {
		return nil, nil, nil, api.NewError(http.StatusConflict, "Duplicate shared scopes in requested Remove", nil)
	}

	existing := sharedScopesToMap(existingSharedScopes)

	toCreate = make([]models.SharedScope, 0)
	toKeep = make([]models.SharedScope, 0)
	toDelete = make([]models.SharedScope, 0)
	existingRequestedToAdd := make([]models.SharedScope, 0)

	for sharedScopeToAdd := range requestedToAdd {
		if existingSharedScope, exists := existing[sharedScopeToAdd]; !exists {
			if sharedScopeToAdd.ScopeType == api.SCOPETYPE_CLUSTER && sharedScopeToAdd.ScopeId != datavolume.ClusterId.String() {
				return nil, nil, nil, api.NewError(http.StatusBadRequest, "Requested cluster scope is not in original datavolume's cluster", nil)
			}
			if sharedScopeToAdd.ScopeType == api.SCOPETYPE_PROJECT && sharedScopeToAdd.ScopeId == datavolume.ProjectId {
				return nil, nil, nil, api.NewError(http.StatusBadRequest, "Cannot share with original project - the datavolume is already shared in the original project by default", nil)
			}

			toCreate = append(toCreate, sharedScopeApiToModel(sharedScopeToAdd, datavolume.ID, tenantId, subjectId))
		} else {
			existingRequestedToAdd = append(existingRequestedToAdd, existingSharedScope)
		}
	}

	for _, existingSharedScope := range existingSharedScopes {
		if _, shouldDelete := requestedToRemove[sharedScopeModelToBase(existingSharedScope)]; shouldDelete {
			toDelete = append(toDelete, existingSharedScope)
		} else {
			toKeep = append(toKeep, existingSharedScope)
		}
	}

	if !areAllScopesPermitted(tenantId, datavolume.ClusterId, toDelete, permittedScopes) {
		return nil, nil, nil, api.NewError(http.StatusForbidden, "Insufficient permissions for toRemove scopes", nil)
	}
	if !areAllScopesPermitted(tenantId, datavolume.ClusterId, existingRequestedToAdd, permittedScopes) {
		// if a user requested to add a scope that already exists, we only need to check if the user has permissions to add it,
		// then we don't need to add it again, of course.
		return nil, nil, nil, api.NewError(http.StatusForbidden, "Insufficient permissions for toAdd scopes", nil)
	}

	err = s.populateScopeNamesAndAncestors(ctx, toCreate, datavolume.ClusterId)
	if err != nil {
		return nil, nil, nil, err
	}

	if !areAllScopesPermitted(tenantId, datavolume.ClusterId, toCreate, permittedScopes) {
		return nil, nil, nil, api.NewError(http.StatusForbidden, "Insufficient permissions for toAdd scopes", nil)
	}

	return toCreate, toKeep, toDelete, nil
}

func (s *DatavolumesApiService) populateScopeNamesAndAncestors(ctx context.Context, sharedScopes []models.SharedScope, clusterId uuid.UUID) *api.Error {
	scopeNames, apiErr := s.orgHierarchyManager.GetScopesNames(ctx, convutils.ConvertSliceTypes(sharedScopes, sharedScopeModelToBase), clusterId.String())
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			// if the scope is not found - convert the returned error to bad request
			apiErr.Code = http.StatusBadRequest
		}
		return apiErr
	}

	for i, sharedScope := range sharedScopes {
		// by this point we should have already validated that the clusterId is the same as the datavolume's clusterId
		apiErr = s.populateAncestorsOfSharedScope(ctx, &sharedScopes[i])
		if apiErr != nil {
			return apiErr
		}

		if scopeName, exists := scopeNames[sharedScopeModelToBase(sharedScope)]; exists {
			sharedScopes[i].ScopeName = scopeName
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get scope name for shared scope <%s/%s> in datavolume <%s>",
				sharedScope.ScopeType, sharedScope.ScopeId, sharedScope.DatavolumeId)
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get scope name of %s/%s",
				sharedScope.ScopeType, sharedScope.ScopeId), nil)
		}
	}

	return nil
}

func (s *DatavolumesApiService) populateAncestorsOfSharedScope(ctx context.Context, sharedScope *models.SharedScope) *api.Error {
	if sharedScope.ScopeType == api.SCOPETYPE_CLUSTER {
		return nil
	}
	if sharedScope.ScopeType == api.SCOPETYPE_DEPARTMENT {
		sharedScope.DepartmentId = sharedScope.ScopeId
		return nil
	}

	// only proceed to get ancestors if the scope is a project
	ancestors, _, err := s.orgHierarchyManager.GetAncestors(ctx, sharedScope.ScopeId, sharedScope.ScopeType)
	if err != nil {
		return api.NewError(http.StatusForbidden, "Insufficient permissions", err)
	}

	if ancestors.DepartmentId != nil {
		sharedScope.DepartmentId = *ancestors.DepartmentId
	}
	sharedScope.ProjectId = sharedScope.ScopeId
	return nil
}

// not used right now but keeping for possible future use.
// for GET datavolumes, get the ancestors of each permitted scope;
// for example: I have READ in a project, a DV was shared with the department of this project, I should see the DV.

//func (s *DatavolumesApiService) getPermittedScopesAncestors(ctx context.Context, permittedScopes *authorization_api.PermittedScopes) (*authorization_api.PermittedScopes, error) {
//	if permittedScopes == nil {
//		return nil, nil
//	}
//
//	departmentPermittedScopes := map[string]bool{}
//	clusterPermittedScopes := map[string]bool{}
//
//	projectsAncestors, err := s.orgHierarchyManager.GetAncestorsMulti(ctx, permittedScopes.Projects, api.SCOPETYPE_PROJECT)
//	if err != nil {
//		return nil, err
//	}
//	departmentsAncestors, err := s.orgHierarchyManager.GetAncestorsMulti(ctx, permittedScopes.Departments, api.SCOPETYPE_DEPARTMENT)
//	if err != nil {
//		return nil, err
//	}
//	scopesAncestors := append(projectsAncestors, departmentsAncestors...)
//
//	for _, ancestors := range scopesAncestors {
//		if ancestors.DepartmentId != nil {
//			departmentPermittedScopes[*ancestors.DepartmentId] = true
//		}
//		if ancestors.ClusterUuid != nil {
//			clusterPermittedScopes[*ancestors.ClusterUuid] = true
//		}
//	}
//
//	for _, permittedDepartment := range permittedScopes.Departments {
//		departmentPermittedScopes[permittedDepartment] = true
//	}
//	for _, permittedCluster := range permittedScopes.Clusters {
//		clusterPermittedScopes[permittedCluster] = true
//	}
//
//	result := &authorization_api.PermittedScopes{
//		Projects:    permittedScopes.Projects,
//		Departments: convertMapToSlice(departmentPermittedScopes),
//		Clusters:    convertMapToSlice(clusterPermittedScopes),
//		Tenants:     permittedScopes.Tenants,
//		System:      permittedScopes.System,
//	}
//
//	return result, nil
//}

func parseConflictOnCreateError(conflictApiErr *api.Error) api.ImplResponse {
	additionalErrorMsg := ""
	if strings.Contains(conflictApiErr.Details, models.DatavolumeClusterNameUniqueConstraintStr) {
		additionalErrorMsg = "datavolume with the same name already exists"
	}
	if strings.Contains(conflictApiErr.Details, models.DatavolumeClusterProjectPvcUniqueConstraintStr) {
		additionalErrorMsg = "datavolume with the same project and pvc name already exists"
	}
	if strings.Contains(conflictApiErr.Details, models.DatavolumeClusterDepartmentPvcUniqueConstraintStr) {
		additionalErrorMsg = "department scope datavolume with the same department and pvc name already exists"
	}
	if strings.Contains(conflictApiErr.Details, models.DatavolumeClusterPvcUniqueConstraintStr) {
		additionalErrorMsg = "cluster scope datavolume with the same cluster and pvc name already exists"
	}

	if additionalErrorMsg == "" {
		return NewErrorResponse(conflictApiErr)
	}
	return NewErrorResponse(api.NewError(http.StatusConflict, fmt.Sprintf("%s: %s", conflictApiErr.Message, additionalErrorMsg), nil))
}

func validateMinimumClusterVersionForDatavolume(ctx context.Context, clusterId uuid.UUID) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfDatavolume)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "Datavolumes are not supported in this cluster version, need minimum version 2.18.0", nil)
	}

	return nil
}

func validateMinimumClusterVersionForClusterAndDepartmentScopeDatavolume(ctx context.Context, clusterId uuid.UUID) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfClusterAndDepartmentScopeDataVolumes)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "Cluster and department scope datavolumes are not supported in this cluster version", nil)
	}

	return nil
}

// validateAndFixCreationRequest validates the creation request and fixes it if necessary and add default values
func ValidateCreationRequestDefaults(datavolumeCreationFields *api.DatavolumeCreationFields) *api.Error {
	if datavolumeCreationFields == nil {
		// Should not reach this point but if it does, return error
		return api.NewError(http.StatusBadRequest, "Invalid request missing datavolume creation fields", nil)
	}

	if (*datavolumeCreationFields).ShouldDeleteOriginalVolume {
		return api.NewError(http.StatusBadRequest, "'shouldDeleteOriginalVolume' is not allowed", nil)
	}

	if (*datavolumeCreationFields).UseOriginalVolume == nil {
		datavolumeCreationFields.UseOriginalVolume = convutils.Ref(true)
	}

	return nil
}
