/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	org_unit_service_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	verbosity "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"
	clients_local "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
)

const (
	LabelKeyRegex           = "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$"
	LabelValueRegex         = "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"
	LabelKeyMaxSize         = 263
	LabelValueMaxSize       = 63
	NodepoolsResourceType   = "nodepools"
	NodepoolAuditEntityType = "node pool"
	NodepoolDefaultName     = "default"
	ExcludedNodepoolName    = "runai-excluded-nodes"
	// DefaultGpuNetworkAccelerationLabelKey Network acceleration - MNNVL Multi Node NVLink
	DefaultGpuNetworkAccelerationLabelKey = "nvidia.com/gpu.clique"
)

// NodePoolsApiService is a service that implements the logic for the NodePoolsApiServicer
// This service should implement the business logic for every endpoint for the NodePoolsApi API.
// Include any external packages or services that will be required by this service.
type NodePoolsApiService struct {
	tenantSettingsClient            clients.TenantSettingsClient
	orgUnitClient                   org_unit_service_client.IOrgUnitClient
	metricsService                  QueryMetricsService
	clusterCapabilityCheckerFactory clients_local.ClusterCapabilityCheckerFactory
}

// NewNodePoolsApiService creates a default api service
func NewNodePoolsApiService(client clients.TenantSettingsClient, orgUnitClient org_unit_service_client.IOrgUnitClient) api.NodePoolsAPIServicer {
	return &NodePoolsApiService{
		tenantSettingsClient:            client,
		orgUnitClient:                   orgUnitClient,
		metricsService:                  NewDefaultMetricsService(),
		clusterCapabilityCheckerFactory: clients_local.LocalClusterCapabilityCheckerFactory,
	}
}

func (s *NodePoolsApiService) GetNodepoolMetrics(ctx context.Context, clusterId string, nodepoolName string,
	start time.Time, end time.Time, metricsType []api.MetricsType, numberOfSamples int32,
) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	measurements := []api.MeasurementResponse{}
	responseCode := http.StatusOK

	for _, metricType := range metricsType {
		result := s.metricsService.QueryNodepoolMetricType(ctx, s.orgUnitClient, clusterId, nodepoolName, start, end, metricType, numberOfSamples)
		if result == nil {
			responseCode = http.StatusMultiStatus
			measurements = append(measurements, api.MeasurementResponse{
				Type: string(metricType),
			})

			continue
		}

		measurement, err := metrics.ParseMetricResultToMeasurement(ctx, string(metricType), *result)
		if err != nil {
			return NewErrorResponse(&api.Error{Code: http.StatusServiceUnavailable, Message: "Error parsing metric response"}), err
		}
		measurements = append(measurements, measurement)
	}

	responseStruct := api.MetricsResponse{Measurements: measurements}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		data := GetMetricsCsvFromMetricsResponse(&responseStruct, metricsType)
		response := runaihttp.CsvFile{Data: data, FileName: "nodepool-metrics.csv"}
		return api.Response(http.StatusOK, response), nil
	}
	return api.Response(responseCode, responseStruct), nil
}

func (s *NodePoolsApiService) CreateNodepool(ctx context.Context, fields api.NodepoolCreateFields) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Create nodepool %s request received", fields.Name)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, fields.ClusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolModel := models.Nodepool{
		Common: models.Common{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:                            fields.Name,
		TenantId:                        tenantId,
		LabelKey:                        &fields.LabelKey,
		LabelValue:                      &fields.LabelValue,
		Phase:                           api.NODEPOOLPHASE_CREATING,
		PhaseMessage:                    nil,
		CreatedBy:                       subjectId,
		UpdatedBy:                       subjectId,
		DeletedBy:                       nil,
		IsDefault:                       false,
		GPUNetworkAccelerationDetection: fields.GpuNetworkAccelerationDetection,
	}
	if fields.GpuNetworkAccelerationLabelKey != nil {
		nodepoolModel.GPUNetworkAccelerationLabelKey = *fields.GpuNetworkAccelerationLabelKey
	}

	ntErr := setNetworkTopology(ctx, tenantId, &nodepoolModel, fields.NetworkTopologyId)
	if ntErr != nil {
		return api.Response(http.StatusBadRequest, ntErr), nil
	}

	cluster, err := dao.GetClusterByIdAndTenantId(ctx, fields.ClusterId, tenantId, verbosity.VerbosityMetadata)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	capChecker := s.clusterCapabilityCheckerFactory(fields.ClusterId)
	validationErr := validateCreateParams(ctx, fields, capChecker)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Create nodepool %s request failed valdiations. error: %s", nodepoolModel.Name, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	if err := setSchedulingConfigurationFieldsOnCreate(ctx, &nodepoolModel, fields, capChecker); err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to apply scheduling configuration on create", err)), nil
	}
	setSchedulingCapabilitiesOnCreateUpdate(ctx, &nodepoolModel, fields.SchedulingCapabilities, true)

	nodepoolSetDefaults(&nodepoolModel)

	validationErr = nodepoolValidations(nodepoolModel)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Create nodepool %s request failed valdiations. error: %v", nodepoolModel.Name, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	nodepoolModel.ClusterId = cluster.ID
	nodepoolModel.ClusterName = cluster.Name

	nodepool, dbErr := dao.CreateNodepool(ctx, nodepoolModel)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Create nodepool %s request failed. error: %v", nodepoolModel.Name, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to create nodepool %s in cluster %s", nodepool.Name, nodepool.ClusterId), dbErr)), nil
	}

	notifyErr := s.orgUnitClient.NodePoolCreated(ctx, cluster.ID.String(), fmt.Sprintf("%d", nodepool.ID), nodepool.Name)
	if notifyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to notify org unit sevice on nodepool %s creation. error: %v", nodepoolModel.Name, notifyErr)
	}

	audit.AuditFromCtx(ctx).WithEntityName(nodepool.Name).WithEntityType(NodepoolAuditEntityType).WithEntityId(strconv.Itoa(int(nodepool.ID)))

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Create nodepool %s request completed successfuly", nodepoolModel.Name)
	apiNp, shapeErr := s.nodepoolToApiResponseWithGpuTypes(ctx, nodepool)
	if shapeErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
	}
	return api.Response(http.StatusCreated, apiNp), nil
}

func (s *NodePoolsApiService) DeleteNodepool(ctx context.Context, nodepoolId string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Delete nodepool with id %s request received", nodepoolId)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepool, dbErr := dao.GetNodepoolById(ctx, tenantId, nodepoolId)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool with id %s", nodepoolId), dbErr)), nil
	}

	validationErr := nodepoolDeleteValidations(nodepool)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Delete nodepool %s with id %s request failed valdiations. error: %v", nodepool.Name, nodepoolId, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	stillCreating := nodepool.Phase == api.NODEPOOLPHASE_CREATING
	if stillCreating {
		dbErr = dao.DeleteNodepool(ctx, tenantId, subjectId, nodepoolId)
	} else {
		dbErr = dao.SetNodepoolPhase(ctx, tenantId, subjectId, nodepoolId, api.NODEPOOLPHASE_DELETING)
	}
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Delete nodepool %s with id %s request failed. error: %v", nodepool.Name, nodepoolId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to set nodepool %s phase to %s", nodepoolId, api.NODEPOOLPHASE_DELETING), dbErr)), nil
	}

	notifyErr := s.orgUnitClient.NodePoolDeleted(ctx, nodepool.ClusterId.String(), nodepoolId, nodepool.Name)
	if notifyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to notify org unit sevice on nodepool %s deletion. error: %v", nodepool.Name, notifyErr)
	}

	audit.AuditFromCtx(ctx).WithEntityName(nodepool.Name).WithEntityType(NodepoolAuditEntityType).WithEntityId(strconv.Itoa(int(nodepool.ID)))

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Delete nodepool %s with id %s request completed successfuly", nodepool.Name, nodepoolId)

	if stillCreating {
		return api.Response(http.StatusNoContent, nil), nil
	}
	return api.Response(http.StatusAccepted, nil), nil
}

func (s *NodePoolsApiService) GetNodepool(ctx context.Context, nodepoolId string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepool, dbErr := dao.GetNodepoolById(ctx, tenantId, nodepoolId)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool with id %s", nodepoolId), dbErr)), nil
	}

	apiNp, shapeErr := s.nodepoolToApiResponseWithGpuTypes(ctx, nodepool)
	if shapeErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
	}

	return api.Response(http.StatusOK, apiNp), nil
}

func (s *NodePoolsApiService) GetNodepoolsMinimal(ctx context.Context, filterBy []string, sortBy api.NodepoolMinimalSortFields, sortOrder string, offset int32, limit int32) (api.ImplResponse, error) {
	nodepools, apiErr := s.getNodepoolsWithFields(ctx, dao.GetNodepoolMinimalFields(), filterBy, sortBy, sortOrder, offset, limit, "nodepools minimal")
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	gpuTypes, apiErr := gpuTypesByNodepool(ctx, nodepools, nil)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	nodepoolsMetadata := make([]api.NodepoolMinimal, len(nodepools))
	for i, nodepool := range nodepools {
		nodepoolsMetadata[i] = nodepool.ToMetadataApiModel(gpuTypesOf(gpuTypes, nodepool))
	}

	return api.Response(http.StatusOK, api.GetNodepoolsMinimal200Response{Nodepools: nodepoolsMetadata}), nil
}

func (s *NodePoolsApiService) GetNodepoolMinimal(ctx context.Context, nodepoolId string) (api.ImplResponse, error) {
	_, scopeErr := GetPermittedClusterScopes(ctx)
	if scopeErr != nil {
		return NewErrorResponse(scopeErr), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		// Should never happen
		return api.Response(http.StatusInternalServerError, "no permitted scopes defined"), nil
	}

	nodepool, apiErr := s.getNodepoolByIdWithFields(ctx, nodepoolId, dao.GetNodepoolMinimalFields())
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	gpuTypes, apiErr := gpuTypesByNodepool(ctx, []models.Nodepool{nodepool}, permittedScopes)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusOK, nodepool.ToMetadataApiModel(gpuTypesOf(gpuTypes, nodepool))), nil
}

func (s *NodePoolsApiService) GetNodepools(ctx context.Context, filterBy []string, sortBy api.NodepoolSortFilterFields, sortOrder string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		// Should never happen
		return api.Response(http.StatusInternalServerError, "no permitted scopes defined"), nil
	}

	filters, validationErr := models.NodepoolQueryParser.Filters(filterBy)
	if validationErr != nil {
		return api.Response(http.StatusBadRequest, validationErr), nil
	}

	sortColumn, sortError := models.NodepoolQueryParser.GetMapping(string(sortBy))
	if sortError != nil {
		return api.Response(http.StatusBadRequest, sortError), nil
	}

	nodepools, dbErr := dao.GetNodepools(ctx, tenantId, *permittedScopes, offset, limit, sortOrder, sortColumn, filters)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Get nodepools request failed. error: %v", dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to list nodepools", dbErr)), nil
	}

	// Older runai-cluster agents call this endpoint with exactly `clusterId==<uuid>` and
	// still read the deprecated top-level placementStrategy field.
	includeDeprecatedFields := isClusterIdOnlyFilter(filters)

	gpuTypes, gpuTypesErr := gpuTypesByNodepool(ctx, nodepools, permittedScopes)
	if gpuTypesErr != nil {
		return NewErrorResponse(gpuTypesErr), nil
	}

	responseNodepools := make([]api.Nodepool, len(nodepools))
	for i, nodepool := range nodepools {
		apiNp, shapeErr := s.nodepoolToApiResponse(ctx, nodepool, gpuTypesOf(gpuTypes, nodepool))
		if shapeErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError,
				fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
		}
		if includeDeprecatedFields {
			populateDeprecatedNodepoolFields(&apiNp, nodepool)
		}
		responseNodepools[i] = apiNp
	}

	return api.Response(http.StatusOK, api.GetNodepools200Response{Nodepools: responseNodepools}), nil
}

func isClusterIdOnlyFilter(filters []queryapi.Filter) bool {
	return len(filters) == 1 &&
		filters[0].Field == models.ClusterIdColumn &&
		filters[0].Operator == "="
}

// populateDeprecatedNodepoolFields fills the deprecated top-level PlacementStrategy on an
// API response object, derived from the canonical fields. Kept solely so older
// runai-cluster agents (that still read it) continue to work.
func populateDeprecatedNodepoolFields(apiNp *api.Nodepool, np models.Nodepool) {
	apiNp.PlacementStrategy = api.NodepoolDeprecatedPlacementStrategy{
		Cpu: nilIfNotValid(np.CpuPlacementStrategy),
		Gpu: nilIfNotValid(np.GpuPlacementStrategy),
	}
}

func (s *NodePoolsApiService) GetNodepoolsMinimalInternal(ctx context.Context, tenantIds []string) (api.ImplResponse, error) {
	nodepools, dbErr := dao.GetNodepoolsForTenants(ctx, tenantIds, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Get nodepools request failed. error: %v", dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to list nodepools", dbErr)), nil
	}

	responseNodepools := make([]api.NodepoolMinimal, len(nodepools))
	for i, nodepool := range nodepools {
		responseNodepools[i] = nodepool.ToMetadataApiModel(nil)
	}

	return api.Response(http.StatusOK, api.GetNodepoolsMinimalInternal200Response{Nodepools: responseNodepools}), nil
}

func (s *NodePoolsApiService) GetNodepoolsInternal(ctx context.Context, clusterUuid string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	nodepools, dbErr := dao.GetNodepoolsForCluster(ctx, fmt.Sprint(tenantId), clusterUuid, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Get nodepools request failed. error: %v", dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to list nodepools", dbErr)), nil
	}

	responseNodepools := make([]api.Nodepool, len(nodepools))
	for i, nodepool := range nodepools {
		apiNp, shapeErr := s.nodepoolToApiResponse(ctx, nodepool, nil)
		if shapeErr != nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError,
				fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
		}
		responseNodepools[i] = apiNp
	}

	return api.Response(http.StatusOK, api.GetNodepoolsInternal200Response{Nodepools: responseNodepools}), nil
}

func (s *NodePoolsApiService) CountNodepools(ctx context.Context, filterBy []string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return api.Response(http.StatusInternalServerError, "Failed to parse tenant id"), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		return api.Response(http.StatusInternalServerError, "no permitted scopes defined"), nil
	}
	filters, filterErr := models.NodepoolQueryParser.Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, filterErr), nil
	}

	count, dbErr := dao.CountNodepools(ctx, tenantId, *permittedScopes, filters)

	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Count nodepools request failed. error: %v", dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to count nodepools", dbErr)), nil
	}

	return api.Response(http.StatusOK, api.CountNodes200Response{Count: count}), nil
}

func (s *NodePoolsApiService) PatchNodepool(ctx context.Context, nodepoolId string, fields api.NodepoolUpdateFields) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Patch nodepool with id %s request received", nodepoolId)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolModel, dbErr := dao.GetNodepoolById(ctx, tenantId, nodepoolId)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool with id %s", nodepoolId), dbErr)), nil
	}
	capChecker := s.clusterCapabilityCheckerFactory(nodepoolModel.ClusterId.String())
	validationErr := validatePatchParams(ctx, nodepoolModel, fields, capChecker)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Patch nodepool %s request failed valdiations. error: %s", nodepoolModel.Name, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	patched, mapErr := mapPatchFieldsToModel(ctx, nodepoolModel, fields, capChecker)
	if mapErr != nil {
		return NewErrorResponse(mapErr), nil
	}

	validationErr = nodepoolValidations(*patched)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Patch nodepool %s with id %s request failed valdiations. err: %v", nodepoolModel.Name, nodepoolId, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	completedResets, apiErr := s.validateSchedulingCapabilitiesChangeAndResetOrgUnits(ctx, nodepoolModel, *patched)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Nodepool %s with id %s capabilities change failed. error: %v", nodepoolModel.Name, nodepoolId, apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	patched.UpdatedBy = subjectId
	if patched.Phase != api.NODEPOOLPHASE_CREATING && clusterFieldsUpdated(*patched, nodepoolModel) {
		patched.Phase = api.NODEPOOLPHASE_UPDATING
		patched.PhaseMessage = nil
	}
	nodepool, dbErr := dao.UpdateNodepool(ctx, *patched)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Patch nodepool %s with id %s request failed. error: %v", nodepoolModel.Name, nodepoolId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to update nodepool %s in cluster %s%s", nodepoolModel.Name, nodepoolModel.ClusterId, describeCompletedResets(completedResets)), dbErr)), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(nodepool.Name).WithEntityType(NodepoolAuditEntityType).WithEntityId(strconv.Itoa(int(nodepool.ID)))

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Patch nodepool %s with id %s request completed successfuly", nodepoolModel.Name, nodepoolId)
	apiNp, shapeErr := s.nodepoolToApiResponseWithGpuTypes(ctx, nodepool)
	if shapeErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
	}
	return api.Response(http.StatusOK, apiNp), nil
}

func (s *NodePoolsApiService) SyncNodepoolsStatus(ctx context.Context, clusterId string, request api.SyncNodepoolsStatusRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Sync nodepools of cluster %s request received", clusterId)
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolsForSync := parseNodepoolsForSync(request)

	synced, failed, dbErr := dao.SyncNodepools(ctx, nodepoolsForSync, tenantId, clusterId, request.DeleteMissing)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync nodepools in cluster %s. error: %v", clusterId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to sync nodepools in cluster %s", clusterId), dbErr)), nil
	}

	if len(failed) != 0 {
		if len(synced) > 0 {
			return api.Response(http.StatusMultiStatus, api.SyncNodepoolsStatus207Response{
				SuccessfullySyncedNodepools: synced,
				FailedNodepools:             failed,
			}), nil
		}

		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to sync nodepools in cluster %s", clusterId), nil)), nil
	}

	return NoContentResponse(), nil
}

func parseNodepoolsForSync(request api.SyncNodepoolsStatusRequest) []models.Nodepool {
	nodepoolsForSync := make([]models.Nodepool, len(request.Nodepools))
	for i := range request.Nodepools {
		nodepool := request.Nodepools[i]

		var phaseMsg *string
		if nodepool.PhaseMessage != "" {
			phaseMsg = &nodepool.PhaseMessage
		}

		nodepoolsForSync[i] = models.Nodepool{
			Name:         nodepool.Name,
			Phase:        nodepool.Phase,
			PhaseMessage: phaseMsg,
			NodepoolStatus: models.NodepoolStatus{
				LabelKey:                       nodepool.Status.LabelKey,
				LabelValue:                     nodepool.Status.LabelValue,
				Nodes:                          nodepool.Status.Nodes,
				GPUNetworkAccelerationDetected: nodepool.Status.GpuNetworkAccelerationDetected,
				Conditions:                     nodepool.Status.Conditions,
			},
		}

		setStatusSchedulingConfigurationFromApiToDao(&nodepoolsForSync[i], &nodepool.Status)
	}

	return nodepoolsForSync
}

func (s *NodePoolsApiService) UpdateNodepool(ctx context.Context, nodepoolId string, fields api.NodepoolUpdateFields) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Update nodepool with id %s request received", nodepoolId)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolModel, dbErr := dao.GetNodepoolById(ctx, tenantId, nodepoolId)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool with id %s", nodepoolId), dbErr)), nil
	}

	capChecker := s.clusterCapabilityCheckerFactory(nodepoolModel.ClusterId.String())
	validationErr := validateUpdateParams(ctx, fields, capChecker)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Update nodepool %s request failed valdiations. error: %s", nodepoolModel.Name, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	updated, ntErr := mapUpdateFieldsToModel(ctx, nodepoolModel, fields, capChecker)
	if ntErr != nil {
		return NewErrorResponse(ntErr), nil
	}
	nodepoolSetDefaults(updated)

	validationErr = nodepoolValidations(*updated)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Update nodepool %s with id %s request failed valdiations. error: %v", nodepoolModel.Name, nodepoolId, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	completedResets, apiErr := s.validateSchedulingCapabilitiesChangeAndResetOrgUnits(ctx, nodepoolModel, *updated)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Nodepool %s with id %s capabilities change failed. error: %v", nodepoolModel.Name, nodepoolId, apiErr.Message)
		return NewErrorResponse(apiErr), nil
	}

	updated.UpdatedBy = subjectId
	if updated.Phase != api.NODEPOOLPHASE_CREATING && clusterFieldsUpdated(*updated, nodepoolModel) {
		updated.Phase = api.NODEPOOLPHASE_UPDATING
		updated.PhaseMessage = nil
	}
	nodepool, dbErr := dao.UpdateNodepool(ctx, *updated)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Update nodepool %s with id %s request failed. error: %v", nodepoolModel.Name, nodepoolId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to update nodepool %s in cluster %s%s", nodepoolModel.Name, nodepoolModel.ClusterId, describeCompletedResets(completedResets)), dbErr)), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(nodepool.Name).WithEntityType(NodepoolAuditEntityType).WithEntityId(strconv.Itoa(int(nodepool.ID)))

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Update nodepool %s with id %s request completed successfuly", nodepoolModel.Name, nodepoolId)
	apiNp, shapeErr := s.nodepoolToApiResponseWithGpuTypes(ctx, nodepool)
	if shapeErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to shape nodepool %s response", nodepool.Name), shapeErr)), nil
	}
	return api.Response(http.StatusOK, apiNp), nil
}

func (s *NodePoolsApiService) GetNodepoolAncestors(ctx context.Context, nodepoolId string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	ancestors, cacheErr := authorization_utils.GetAncestorsCache(ctx, NodepoolsResourceType, nodepoolId)
	if cacheErr == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find nodepoolId <%s> ancestors in cache. err: %v, continuing", nodepoolId, cacheErr)

	nodepool, dbErr := dao.GetNodepoolById(ctx, tenantId, nodepoolId)
	if dbErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool with id %s", nodepoolId), dbErr)), nil
	}

	ancestors = &authorization_utils.Ancestors{
		TenantId:    fmt.Sprintf("%d", nodepool.TenantId),
		ClusterUuid: ptr.To(nodepool.ClusterId.String()),
	}
	authorization_utils.SetAncestorsCache(ctx, NodepoolsResourceType, nodepoolId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *NodePoolsApiService) ClusterNotification(ctx context.Context, clusterNotification api.ClusterNotification) (api.ImplResponse, error) {
	switch clusterNotification.Operation {
	case api.NOTIFICATIONOPERATION_CREATE:
		return s.createDefaultNodepool(ctx, clusterNotification)
	case api.NOTIFICATIONOPERATION_DELETE:
		return deleteClusterNodepools(ctx, clusterNotification)
	}

	return api.Response(http.StatusBadRequest, fmt.Sprintf("Unsupported notification operation %s", clusterNotification.Operation)), nil
}

func (s *NodePoolsApiService) createDefaultNodepool(ctx context.Context, clusterNotification api.ClusterNotification) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Create default nodepool request received")
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolModel := models.Nodepool{
		Common: models.Common{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:                       NodepoolDefaultName,
		TenantId:                   tenantId,
		LabelKey:                   nil,
		LabelValue:                 nil,
		Phase:                      api.NODEPOOLPHASE_CREATING,
		PhaseMessage:               nil,
		CreatedBy:                  subjectId,
		UpdatedBy:                  subjectId,
		DeletedBy:                  nil,
		IsDefault:                  true,
		GpuPlacementStrategy:       "",
		GpuDevicePlacementStrategy: "",
		CpuPlacementStrategy:       "",
		MinGuaranteedRuntime:       nil,
	}
	setSchedulingCapabilitiesOnCreateUpdate(ctx, &nodepoolModel, clusterNotification.SchedulingCapabilities, true)
	nodepoolSetDefaults(&nodepoolModel)

	nodepoolModel.ClusterId = uuid.MustParse(clusterNotification.Id)
	nodepoolModel.ClusterName = clusterNotification.Name

	nodepoolModel, dbErr := dao.CreateNodepool(ctx, nodepoolModel)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Create nodepool %s request failed. error: %v", nodepoolModel.Name, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to create nodepool %s in cluster %s", nodepoolModel.Name, nodepoolModel.ClusterId), dbErr)), nil
	}

	notifyErr := s.orgUnitClient.DefaultNodePoolCreated(ctx, clusterNotification.Id, fmt.Sprintf("%d", nodepoolModel.ID))
	if notifyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to notify org unit sevice on default nodepool creation of cluster %s. error: %v", nodepoolModel.ClusterId, notifyErr)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Create default nodepool for cluster %s request completed successfuly", clusterNotification.Name)
	return api.Response(http.StatusNoContent, nil), nil
}

func deleteClusterNodepools(ctx context.Context, clusterNotification api.ClusterNotification) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("delete cluster %s nodepools request received", clusterNotification.Id)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	dbErr := dao.DeleteClusterNodepools(ctx, tenantId, subjectId, clusterNotification.Id)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("delete cluster %s nodepools request failed. error: %v", clusterNotification.Id, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to delete cluster %s nodepools", clusterNotification.Id), dbErr)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("delete cluster %s nodepools request completed successfuly", clusterNotification.Id)
	return api.Response(http.StatusNoContent, nil), nil
}

func mapUpdateFieldsToModel(ctx context.Context, nodepoolModel models.Nodepool, updateFields api.NodepoolUpdateFields, capChecker clients_local.ClusterCapabilityChecker) (*models.Nodepool, *api.Error) {
	result := &models.Nodepool{}
	(&nodepoolModel).DeepCopyInto(result)

	result.LabelKey = updateFields.LabelKey
	result.LabelValue = updateFields.LabelValue

	result.GPUNetworkAccelerationDetection = ptr.To(convutils.Deref(updateFields.GpuNetworkAccelerationDetection, api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO))
	result.GPUNetworkAccelerationLabelKey = convutils.Deref(updateFields.GpuNetworkAccelerationLabelKey, "")

	if err := setSchedulingConfigurationFieldsOnUpdate(ctx, result, updateFields, capChecker); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "Failed to apply scheduling configuration on update", err)
	}
	setSchedulingCapabilitiesOnCreateUpdate(ctx, result, updateFields.SchedulingCapabilities, false)

	ntErr := setNetworkTopology(ctx, nodepoolModel.TenantId, result, updateFields.NetworkTopologyId)
	if ntErr != nil {
		return nil, ntErr
	}
	return result, nil
}

func mapPatchFieldsToModel(ctx context.Context, nodepoolModel models.Nodepool, patchFields api.NodepoolUpdateFields, capChecker clients_local.ClusterCapabilityChecker) (*models.Nodepool, *api.Error) {
	result := &models.Nodepool{}
	(&nodepoolModel).DeepCopyInto(result)

	if patchFields.LabelKey != nil {
		result.LabelKey = patchFields.LabelKey
	}
	if patchFields.LabelValue != nil {
		result.LabelValue = patchFields.LabelValue
	}

	if patchFields.GpuNetworkAccelerationDetection != nil {
		result.GPUNetworkAccelerationDetection = ptr.To(convutils.Deref(patchFields.GpuNetworkAccelerationDetection, api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO))
	}

	if patchFields.GpuNetworkAccelerationLabelKey != nil {
		result.GPUNetworkAccelerationLabelKey = *patchFields.GpuNetworkAccelerationLabelKey
	}

	if err := setSchedulingConfigurationFieldsOnPatch(ctx, result, patchFields, capChecker); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, err.Error(), err)
	}
	setSchedulingCapabilitiesOnPatch(result, patchFields.SchedulingCapabilities)

	ntErr := setNetworkTopology(ctx, nodepoolModel.TenantId, result, patchFields.NetworkTopologyId)
	if ntErr != nil {
		return nil, ntErr
	}
	return result, nil
}

// nodepoolToApiResponse builds the API response object for a nodepool, applying
// cluster-capability-based response shaping. All endpoints that return a Nodepool
// should funnel through this method so unsupported fields are stripped consistently.
func (s *NodePoolsApiService) nodepoolToApiResponse(
	ctx context.Context, nodepool models.Nodepool, gpuTypes []api.NodepoolGpuTypeDetails,
) (api.Nodepool, error) {
	apiNp := nodepoolDaoToApi(nodepool)
	apiNp.GpuTypes = gpuTypes
	capChecker := s.clusterCapabilityCheckerFactory(nodepool.ClusterId.String())
	if err := stripUnsupportedSchedulingFields(ctx, apiNp.SchedulingConfiguration, capChecker); err != nil {
		return api.Nodepool{}, err
	}
	return apiNp, nil
}

func (s *NodePoolsApiService) nodepoolToApiResponseWithGpuTypes(
	ctx context.Context, nodepool models.Nodepool,
) (api.Nodepool, error) {
	gpuTypes, apiErr := gpuTypesByNodepool(ctx, []models.Nodepool{nodepool}, nil)
	if apiErr != nil {
		return api.Nodepool{}, apiErr
	}
	return s.nodepoolToApiResponse(ctx, nodepool, gpuTypesOf(gpuTypes, nodepool))
}

func gpuTypesByNodepool(
	ctx context.Context, nodepools []models.Nodepool, permittedScopes *authorization_api.PermittedScopes,
) (map[dao.NodepoolKey][]api.NodepoolGpuTypeDetails, *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return nil, err
	}

	if permittedScopes == nil {
		scopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
		if !ok || scopes == nil {
			// Should never happen
			return nil, api.NewError(http.StatusInternalServerError, "no permitted scopes defined", nil)
		}
		permittedScopes = scopes
	}

	keys := make([]dao.NodepoolKey, len(nodepools))
	for i, nodepool := range nodepools {
		keys[i] = dao.NewNodepoolKey(nodepool.ClusterId.String(), nodepool.Name)
	}

	return dao.GetNodepoolsGpuTypes(ctx, tenantId, *permittedScopes, keys)
}

// gpuTypesOf picks one nodepool's entry out of a gpuTypesByNodepool result.
func gpuTypesOf(
	gpuTypes map[dao.NodepoolKey][]api.NodepoolGpuTypeDetails, nodepool models.Nodepool,
) []api.NodepoolGpuTypeDetails {
	return gpuTypes[dao.NewNodepoolKey(nodepool.ClusterId.String(), nodepool.Name)]
}

func nodepoolDaoToApi(nodepool models.Nodepool) api.Nodepool {
	apiNodepool := api.Nodepool{
		Name:                            nodepool.Name,
		LabelKey:                        convutils.Deref(nodepool.LabelKey, ""),
		LabelValue:                      convutils.Deref(nodepool.LabelValue, ""),
		ClusterId:                       nodepool.ClusterId.String(),
		Phase:                           nodepool.Phase,
		PhaseMessage:                    convutils.Deref(nodepool.PhaseMessage, ""),
		Id:                              fmt.Sprintf("%d", nodepool.ID),
		TenantId:                        nodepool.TenantId,
		ClusterName:                     nodepool.ClusterName,
		CreatedBy:                       nodepool.CreatedBy,
		CreatedAt:                       nodepool.CreatedAt,
		UpdatedBy:                       nodepool.UpdatedBy,
		UpdatedAt:                       nodepool.UpdatedAt,
		DeletedBy:                       convutils.Deref(nodepool.DeletedBy, ""),
		DeletedAt:                       dbmodels.GormDeletedAtToNullableTime(nodepool.DeletedAt),
		IsDefault:                       nodepool.IsDefault,
		Status:                          getApiStatus(nodepool),
		GpuNetworkAccelerationDetection: nodepool.GPUNetworkAccelerationDetection,
		GpuNetworkAccelerationLabelKey:  nodepool.GPUNetworkAccelerationLabelKey,

		SchedulingConfiguration: nodepool.GetSchedulingConfiguration(),

		SchedulingCapabilities: nodepool.GetSchedulingCapabilities(),
	}

	if nodepool.NetworkTopologyId != nil {
		apiNodepool.NetworkTopologyId = convutils.Ref(nodepool.NetworkTopologyId.String())
	}
	apiNodepool.NetworkTopologyName = nodepool.NetworkTopologyName

	return apiNodepool
}

func nilIfNotValid(strategy api.PlacementStrategy) *api.PlacementStrategy {
	if !strategy.IsValid() {
		return nil
	}
	return ptr.To(strategy)
}

func getApiStatus(dbNodepool models.Nodepool) api.NodepoolSyncFieldsStatus {
	status := dbNodepool.NodepoolStatus
	result := api.NodepoolSyncFieldsStatus{
		LabelKey:   status.LabelKey,
		LabelValue: status.LabelValue,
		Nodes: func() []string {
			nodes, _ := parseNodeNames(status.Nodes)
			return nodes
		}(),
		GpuNetworkAccelerationDetected: status.GPUNetworkAccelerationDetected,
		// The deprecated top-level PlacementStrategy remains in NodepoolSyncUpdateFields only to
		// accept syncs from old-versioned clusters; we deliberately do not echo it back to API
		// consumers — the canonical fields below are the source of truth.

		SchedulingConfiguration: &api.SchedulingConfiguration{
			MinGuaranteedRuntime: status.MinGuaranteedRuntime,
			GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
				Node:   nilIfNotValid(status.GpuPlacementStrategy),
				Device: nilIfNotValid(status.GpuDevicePlacementStrategy),
			},
			CpuPodsPlacementStrategy: &api.CpuPodsPlacementStrategy{
				Node: nilIfNotValid(status.CpuPlacementStrategy),
			},
			PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
				Cpu: nilIfNotValid(status.CpuPlacementStrategy),
				Gpu: nilIfNotValid(status.GpuPlacementStrategy),
			},
		},
		Conditions: status.Conditions,
	}

	return result
}

func parseNodeNames(nodes []string) ([]string, error) {
	inputVal := reflect.ValueOf(nodes)

	var output []string
	for i := 0; i < inputVal.Len(); i++ {
		item := inputVal.Index(i).String()
		item = strings.NewReplacer("\\\"", "\"", "[", "", "]", "", "\"", "").Replace(item)
		output = append(output, strings.TrimSpace(item))
	}

	return output, nil
}

func nodepoolSetDefaults(fields *models.Nodepool) {
	if fields.GpuPlacementStrategy == "" {
		fields.GpuPlacementStrategy = api.PLACEMENTSTRATEGY_BINPACK
	}

	if fields.GpuDevicePlacementStrategy == "" {
		fields.GpuDevicePlacementStrategy = api.PLACEMENTSTRATEGY_BINPACK
	}

	if fields.CpuPlacementStrategy == "" {
		fields.CpuPlacementStrategy = api.PLACEMENTSTRATEGY_BINPACK
	}

	if fields.GPUNetworkAccelerationDetection == nil {
		fields.GPUNetworkAccelerationDetection = ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO)
	}

	if fields.GPUNetworkAccelerationLabelKey == "" {
		fields.GPUNetworkAccelerationLabelKey = DefaultGpuNetworkAccelerationLabelKey
	}
}

func gpuNetworkAccelerationDetectionValidation(
	ctx context.Context,
	gpuNetworkAccelerationDetection *api.NodepoolGpuNetworkAccelerationDetection,
	capChecker clients_local.ClusterCapabilityChecker,
) error {
	if gpuNetworkAccelerationDetection == nil {
		return nil
	}

	clusterSupported, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfGPUNetworkAccelerationDetection)
	if err != nil {
		return fmt.Errorf("failed to check cluster support for GPUNetworkAccelerationDetection. err: %v", err)
	}
	if !clusterSupported {
		return fmt.Errorf("GPU Network Acceleration detection supported from version %s", clients_local.LocalMinVersionForSupporting(clients_local.LocalMinVersionOfGPUNetworkAccelerationDetection))
	}
	return nil
}

func nodepoolValidations(nodepool models.Nodepool) error {
	err := validateLabels(nodepool.LabelKey, nodepool.LabelValue, nodepool.IsDefault)
	if err != nil {
		return err
	}
	err = validateName(nodepool)
	if err != nil {
		return err
	}
	err = validateSchedulingMode(nodepool)
	if err != nil {
		return err
	}
	return nil
}

func validateName(nodepool models.Nodepool) error {
	if !nodepool.IsDefault && nodepool.Name == NodepoolDefaultName {
		return fmt.Errorf("can not create nodepool named %s which is not the system default nodepool", NodepoolDefaultName)
	}
	if nodepool.IsDefault && nodepool.Name != NodepoolDefaultName {
		return fmt.Errorf("can not create system default nodepool which it name is not %s", NodepoolDefaultName)
	}
	if nodepool.Name == ExcludedNodepoolName {
		return fmt.Errorf("can not create a nodepool named %s, which is a system reserved name", ExcludedNodepoolName)
	}

	return nil
}

func validateLabels(key *string, value *string, isDefault bool) error {
	if isDefault && (key != nil || value != nil) {
		return fmt.Errorf("default node pool should not have label key or value")
	}
	if !isDefault && (key == nil || value == nil) {
		return fmt.Errorf("node pool should have label key and value")
	}

	if key != nil {
		labelKeyRegex := regexp.MustCompile(LabelKeyRegex)
		if !labelKeyRegex.MatchString(*key) {
			return fmt.Errorf("nodepool label key must compile regular expression %s", LabelKeyRegex)
		}
		if len(*key) > LabelKeyMaxSize {
			return fmt.Errorf("nodepool label key length should be less than %d", LabelKeyMaxSize)
		}
	}

	if value != nil {
		labelValueRegex := regexp.MustCompile(LabelValueRegex)
		if !labelValueRegex.MatchString(*value) {
			return fmt.Errorf("nodepool label value must compile regular expression %s", LabelValueRegex)
		}
		if len(*value) > LabelValueMaxSize {
			return fmt.Errorf("nodepool label value length should be less than %d", LabelValueMaxSize)
		}
	}

	return nil
}

func validateSchedulingMode(nodepool models.Nodepool) error {
	if nodepool.AllowCpuQuota && !nodepool.AllowQuota {
		return fmt.Errorf("SchedulingCapabilities.CpuQuota can only be enabled when SchedulingCapabilities.Quota is enabled")
	}
	return nil
}

func nodepoolDeleteValidations(nodepool models.Nodepool) error {
	if nodepool.IsDefault || nodepool.Name == NodepoolDefaultName {
		return fmt.Errorf("can not delete default nodepool")
	}
	if nodepool.Phase == api.NODEPOOLPHASE_DELETING || nodepool.Phase == api.NODEPOOLPHASE_DELETED {
		return fmt.Errorf("nodepool is already deleted")
	}

	return nil
}

// getNodepoolByIdWithFields is a helper function to retrieve a nodepool by ID with specific fields
func (s *NodePoolsApiService) getNodepoolByIdWithFields(ctx context.Context, nodepoolId string, selectFields []string) (models.Nodepool, *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return models.Nodepool{}, err
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		return models.Nodepool{}, api.NewError(http.StatusInternalServerError, "no permitted scopes defined", nil)
	}

	nodepool, dbErr := dao.GetNodepoolMinimalById(ctx, tenantId, selectFields, nodepoolId, *permittedScopes)
	if dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return models.Nodepool{}, api.NewError(http.StatusNotFound,
				fmt.Sprintf("Nodepool with id %s not found", nodepoolId), dbErr)
		}
		return models.Nodepool{}, api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get nodepool metadata with id %s", nodepoolId), dbErr)
	}

	return nodepool, nil
}

// getNodepoolsWithFields is a helper function to retrieve nodepools with specific fields
func (s *NodePoolsApiService) getNodepoolsWithFields(ctx context.Context, selectFields []string,
	filterBy []string,
	sortBy api.NodepoolMinimalSortFields,
	sortOrder string,
	offset int32,
	limit int32,
	errorContext string,
) ([]models.Nodepool, *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return nil, err
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		return nil, api.NewError(http.StatusInternalServerError, "no permitted scopes defined", nil)
	}

	filters, validationErr := models.NodepoolMetadataQueryParser.Filters(filterBy)
	if validationErr != nil {
		return nil, api.NewError(http.StatusBadRequest, validationErr.Error(), validationErr)
	}

	sortColumn, sortError := models.NodepoolMetadataQueryParser.GetMapping(string(sortBy))
	if sortError != nil {
		return nil, api.NewError(http.StatusBadRequest, sortError.Error(), sortError)
	}

	nodepools, dbErr := dao.GetNodepoolsWithFields(ctx, tenantId, selectFields, offset, limit, sortOrder, sortColumn, filters, *permittedScopes)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s request failed. error: %v", errorContext, dbErr)
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to list %s", errorContext), dbErr)
	}

	return nodepools, nil
}

// LegacyDeleteNodepool - Delete nodepool (legacy)
// Deprecated: delegates to DeleteNodepool, clusterId is ignored as the nodepool is identified by nodepoolId.
func (s *NodePoolsApiService) LegacyDeleteNodepool(ctx context.Context, clusterId string, nodepoolId string) (api.ImplResponse, error) {
	return s.DeleteNodepool(ctx, nodepoolId)
}

// LegacyUpdateNodepool - Update nodepool (legacy)
// Deprecated: delegates to UpdateNodepool, clusterId is ignored as the nodepool is identified by nodepoolId.
func (s *NodePoolsApiService) LegacyUpdateNodepool(ctx context.Context, clusterId string, nodepoolId string, req api.LegacyUpdateNodePoolRequest) (api.ImplResponse, error) {
	// Translate the legacy top-level PlacementStrategy into schedulingConfiguration.
	var placementStrategy api.SchedulingConfigurationPlacementStrategy
	if req.PlacementStrategy.Cpu != "" {
		cpu := api.PlacementStrategy(req.PlacementStrategy.Cpu)
		placementStrategy.Cpu = &cpu
	}
	if req.PlacementStrategy.Gpu != "" {
		gpu := api.PlacementStrategy(req.PlacementStrategy.Gpu)
		placementStrategy.Gpu = &gpu
	}

	updateFields := api.NodepoolUpdateFields{
		LabelKey:                &req.LabelKey,
		LabelValue:              &req.LabelValue,
		SchedulingConfiguration: &api.SchedulingConfiguration{PlacementStrategy: placementStrategy},
	}
	return s.UpdateNodepool(ctx, nodepoolId, updateFields)
}

// LegacyUpdateNodepoolLabels - Update nodepool labels (legacy)
// Deprecated: delegates to PatchNodepool, clusterId is ignored as the nodepool is identified by nodepoolId.
func (s *NodePoolsApiService) LegacyUpdateNodepoolLabels(ctx context.Context, clusterId string, nodepoolId string, req api.LegacyNodePoolLabelsRequest) (api.ImplResponse, error) {
	updateFields := api.NodepoolUpdateFields{
		LabelKey:   &req.LabelKey,
		LabelValue: &req.LabelValue,
	}
	return s.PatchNodepool(ctx, nodepoolId, updateFields)
}

// LegacyGetNodepools - Get cluster's node pools (legacy)
// Deprecated: returns node pools for a cluster in the legacy format.
func (s *NodePoolsApiService) LegacyGetNodepools(ctx context.Context, clusterId string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Legacy get nodepools for cluster %s request received", clusterId)
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepools, dbErr := dao.GetNodepoolsForCluster(ctx, fmt.Sprint(tenantId), clusterId, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Legacy get nodepools for cluster %s failed. error: %v", clusterId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to list nodepools", dbErr)), nil
	}

	legacyNodepools := make([]api.LegacyNodePool, len(nodepools))
	for i, nodepool := range nodepools {
		legacyNodepools[i] = nodepoolDaoToLegacyApi(nodepool)
	}

	return api.Response(http.StatusOK, legacyNodepools), nil
}

// LegacyCreateNodepool - Create a node pool (legacy)
// Deprecated: delegates to CreateNodepool with legacy request format.
func (s *NodePoolsApiService) LegacyCreateNodepool(ctx context.Context, clusterId string, req api.LegacyNodePoolCreateRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Legacy create nodepool %s for cluster %s request received", req.Name, clusterId)

	createFields := api.NodepoolCreateFields{
		Name:       req.Name,
		ClusterId:  clusterId,
		LabelKey:   req.LabelKey,
		LabelValue: req.LabelValue,
	}

	resp, err := s.CreateNodepool(ctx, createFields)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// LegacyUpdateNodepoolStatus - Update node pool status (legacy)
// Deprecated: updates node pool status using the legacy format.
func (s *NodePoolsApiService) LegacyUpdateNodepoolStatus(ctx context.Context, clusterId string, name string, req api.LegacyNodePoolStatusRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Legacy update nodepool status for cluster %s, name %s request received", clusterId, name)
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	phase := api.NodepoolPhase(req.Status)
	var phaseMsg *string
	if req.StatusMessage != "" {
		phaseMsg = &req.StatusMessage
	}

	nodepoolsForSync := []models.Nodepool{
		{
			Name:         name,
			Phase:        phase,
			PhaseMessage: phaseMsg,
			NodepoolStatus: models.NodepoolStatus{
				Nodes: parseLegacyNodes(req.Nodes),
			},
		},
	}

	_, failed, dbErr := dao.SyncNodepools(ctx, nodepoolsForSync, tenantId, clusterId, false)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Legacy update nodepool status for cluster %s failed. error: %v", clusterId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to update nodepool status in cluster %s", clusterId), dbErr)), nil
	}

	if len(failed) > 0 {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to update nodepool %s status in cluster %s", name, clusterId), nil)), nil
	}

	return api.Response(http.StatusCreated, nil), nil
}

// LegacySyncNodepools - Sync node pools (legacy)
// Deprecated: syncs node pool statuses using the legacy format.
func (s *NodePoolsApiService) LegacySyncNodepools(ctx context.Context, clusterId string, reqs []api.LegacyNodePoolStatusRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Legacy sync nodepools for cluster %s request received", clusterId)
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodepoolsForSync := make([]models.Nodepool, 0, len(reqs))
	for _, req := range reqs {
		if req.Status == "" {
			continue
		}
		phase := api.NodepoolPhase(req.Status)
		var phaseMsg *string
		if req.StatusMessage != "" {
			phaseMsg = &req.StatusMessage
		}
		nodepoolsForSync = append(nodepoolsForSync, models.Nodepool{
			Phase:        phase,
			PhaseMessage: phaseMsg,
			NodepoolStatus: models.NodepoolStatus{
				Nodes: parseLegacyNodes(req.Nodes),
			},
		})
	}

	_, _, dbErr := dao.SyncNodepools(ctx, nodepoolsForSync, tenantId, clusterId, true)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Legacy sync nodepools for cluster %s failed. error: %v", clusterId, dbErr)
		return NewErrorResponse(api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to sync nodepools in cluster %s", clusterId), dbErr)), nil
	}

	return api.Response(http.StatusOK, nil), nil
}

// parseLegacyNodes parses the legacy nodes string (JSON-encoded or comma-separated) into a string slice.
func parseLegacyNodes(nodes string) []string {
	if nodes == "" {
		return nil
	}
	// Try JSON array first
	var parsed []string
	if err := json.Unmarshal([]byte(nodes), &parsed); err == nil {
		return parsed
	}
	// Fall back to comma-separated
	return strings.Split(nodes, ",")
}

func nodepoolDaoToLegacyApi(nodepool models.Nodepool) api.LegacyNodePool {
	var statusStr string
	switch nodepool.Phase {
	case api.NODEPOOLPHASE_CREATING:
		statusStr = "Creating"
	case api.NODEPOOLPHASE_UPDATING:
		statusStr = "Updating"
	case api.NODEPOOLPHASE_DELETING:
		statusStr = "Deleting"
	case api.NODEPOOLPHASE_EMPTY:
		statusStr = "Empty"
	case api.NODEPOOLPHASE_UNSCHEDULABLE:
		statusStr = "Unschedulable"
	case api.NODEPOOLPHASE_READY:
		statusStr = "Ready"
	case api.NODEPOOLPHASE_DELETED:
		statusStr = "Deleted"
	default:
		statusStr = string(nodepool.Phase)
	}

	var nodesStr string
	if len(nodepool.Nodes) > 0 {
		nodesBytes, _ := json.Marshal(nodepool.Nodes)
		nodesStr = string(nodesBytes)
	}

	return api.LegacyNodePool{
		Id:         float32(nodepool.ID),
		Name:       nodepool.Name,
		LabelKey:   convutils.Deref(nodepool.LabelKey, ""),
		LabelValue: convutils.Deref(nodepool.LabelValue, ""),
		PlacementStrategy: api.LegacyPlacementStrategy{
			Cpu: string(nodepool.CpuPlacementStrategy),
			Gpu: string(nodepool.GpuPlacementStrategy),
		},
		ClusterId: nodepool.ClusterId.String(),
		CreatedAt: nodepool.CreatedAt,
		UpdatedAt: nodepool.UpdatedAt,
		DeletedAt: func() time.Time {
			t := dbmodels.GormDeletedAtToNullableTime(nodepool.DeletedAt)
			if t != nil {
				return *t
			}
			return time.Time{}
		}(),
		Status:        statusStr,
		StatusMessage: convutils.Deref(nodepool.PhaseMessage, ""),
		Nodes:         nodesStr,
		CreatedBy:     nodepool.CreatedBy,
		UpdatedBy:     nodepool.UpdatedBy,
		IsDefault:     nodepool.IsDefault,
	}
}
