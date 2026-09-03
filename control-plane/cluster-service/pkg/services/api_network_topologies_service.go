/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/adapters"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

const (
	NetworkTopologiesAuditEntityType = "network topology"
)

// NetworkTopologiesApiService is a service that implements the logic for the NetworkTopologiesApiServicer
// This service should implement the business logic for every endpoint for the NetworkTopologiesApi API.
// Include any external packages or services that will be required by this service.
type NetworkTopologiesApiService struct{}

// NewNetworkTopologiesApiService creates a default api service
func NewNetworkTopologiesApiService() api.NetworkTopologiesAPIServicer {
	return &NetworkTopologiesApiService{}
}

func (s *NetworkTopologiesApiService) SyncNetworkTopologies(ctx context.Context, clusterId string, request api.NetworkTopologySyncRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Syncing network topologies for: %+v", request)
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Failed to parse cluster uuid", err)), nil
	}

	tenantId, parseErr := utils.ParseTenantId(ctx)
	if parseErr != nil {
		errMessage := fmt.Sprintf("parse tenant id error: %v", parseErr)
		logging.Errorf("%s", errMessage)
		return NewErrorResponse(api.NewError(http.StatusBadRequest, errMessage, parseErr)), nil
	}

	subjectId := convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "Cluster")
	apiErr := dao.SyncNetworkTopologies(ctx, tenantId, subjectId, clusterUuid, request)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync network topologies. err: %v", apiErr)
		return NewErrorResponse(apiErr), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *NetworkTopologiesApiService) GetNetworkTopologies(ctx context.Context, filterBy []string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		return api.Response(http.StatusInternalServerError, "no permitted scopes defined"), nil
	}

	filters, filterErr := models.NetworkTopologyQueryParser.Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, filterErr), nil
	}
	// Fetch from the view to include nodePools
	networkTopologiesView, dbErr := dao.GetNetworkTopologiesView(ctx, tenantId, permittedScopes, filters)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}
	responseNetworkTopologies := make([]api.NetworkTopology, len(networkTopologiesView))
	for i, networkTopologyView := range networkTopologiesView {
		apiModel := networkTopologyViewDaoToApi(networkTopologyView)
		responseNetworkTopologies[i] = apiModel
	}

	return api.Response(http.StatusOK, api.GetNetworkTopologies200Response{NetworkTopologies: responseNetworkTopologies}), nil
}

func (s *NetworkTopologiesApiService) CreateNetworkTopology(ctx context.Context, networkTopologyCreateRequest api.NetworkTopologyCreateRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Create network topology request received with name <%s>", networkTopologyCreateRequest.Name)
	tenantId, subjectId, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	clusterId, parseError := uuid.Parse(networkTopologyCreateRequest.ClusterId)
	if parseError != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Cluster Id should be in a valid UUID format", parseError)), nil
	}

	isEmptyLevels := len(networkTopologyCreateRequest.Levels) == 0
	if isEmptyLevels {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "Levels can not be empty", nil)), nil
	}

	// Validate cluster exists and user has access
	_, err = ValidatePermittedScope(ctx, tenantId, networkTopologyCreateRequest.ClusterId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	networkTopologyModel := createRequestToNetworkTopology(subjectId, networkTopologyCreateRequest, clusterId, tenantId)

	createdNetworkTopology, dbErr := dao.CreateNetworkTopology(ctx, networkTopologyModel)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	audit.AuditFromCtx(ctx).WithEntityName(createdNetworkTopology.Name).WithEntityType(NetworkTopologiesAuditEntityType).WithEntityId(createdNetworkTopology.ID.String())
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Create network topology %s request completed successfuly", createdNetworkTopology.Name)

	responseNetworkTopology := networkTopologyDaoToApi(*createdNetworkTopology)
	return api.Response(http.StatusCreated, responseNetworkTopology), nil
}

func createRequestToNetworkTopology(subjectId string,
	networkTopologyCreateRequest api.NetworkTopologyCreateRequest,
	clusterId uuid.UUID, tenantId int32,
) models.NetworkTopology {
	now := time.Now()
	return models.NetworkTopology{
		CreatedAt: now,
		CreatedBy: &subjectId,
		Name:      networkTopologyCreateRequest.Name,
		Levels:    networkTopologyCreateRequest.Levels,
		Source:    dao.NetworkTopologiesRunAiSource,
		ClusterID: clusterId,
		TenantID:  tenantId,
		Phase:     models.NetworkTopologyPhaseCreating,
	}
}

func (s *NetworkTopologiesApiService) GetNetworkTopologiesCount(ctx context.Context, filterBy []string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return api.Response(http.StatusInternalServerError, "Failed to parse tenant id"), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || permittedScopes == nil {
		return api.Response(http.StatusInternalServerError, "no permitted scopes defined"), nil
	}

	filters, filterErr := models.NetworkTopologyQueryParser.Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, filterErr), nil
	}

	count, dbErr := dao.CountNetworkTopologies(ctx, tenantId, *permittedScopes, filters)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	return api.Response(http.StatusOK, api.CountResponse{Count: count}), nil
}

func (s *NetworkTopologiesApiService) GetNetworkTopologyById(ctx context.Context, networkTopologyId string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	validationError := uuid.Validate(networkTopologyId)
	if validationError != nil {
		return api.Response(http.StatusBadRequest, validationError), nil
	}
	networkTopologyView, dbErr := dao.GetNetworkTopologyViewById(ctx, tenantId, networkTopologyId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	// Validate cluster access
	_, err = ValidatePermittedScope(ctx, tenantId, networkTopologyView.ClusterID.String())
	if err != nil {
		return NewErrorResponse(err), nil
	}
	responseNetworkTopology := networkTopologyViewDaoToApi(*networkTopologyView)
	return api.Response(http.StatusOK, responseNetworkTopology), nil
}

func (s *NetworkTopologiesApiService) DeleteNetworkTopology(ctx context.Context, networkTopologyId string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Delete network topology with id %s request received", networkTopologyId)
	tenantId, _, err := ParseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	validationError := uuid.Validate(networkTopologyId)
	if validationError != nil {
		return api.Response(http.StatusBadRequest, validationError), nil
	}

	networkTopologyModel, dbErr := dao.GetNetworkTopologyById(ctx, tenantId, networkTopologyId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	if networkTopologyModel.Source != dao.NetworkTopologiesRunAiSource {
		errMessage := fmt.Sprintf("Cannot delete network topology with source '%s'. Only network topologies with source '%s' can be deleted via API",
			networkTopologyModel.Source, dao.NetworkTopologiesRunAiSource)
		responseErr := api.NewError(http.StatusBadRequest, errMessage, nil)
		return NewErrorResponse(responseErr), nil
	}
	// Validate cluster access
	_, err = ValidatePermittedScope(ctx, tenantId, networkTopologyModel.ClusterID.String())
	if err != nil {
		return NewErrorResponse(err), nil
	}

	validationErr := networkTopologyDeleteValidations(*networkTopologyModel)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Delete network topology %s request failed validations. error: %s", networkTopologyModel.Name, validationErr.Error())
		return api.Response(http.StatusBadRequest, validationErr.Error()), nil
	}

	dbErr = dao.DeleteNetworkTopology(ctx, tenantId, networkTopologyModel.ClusterID, networkTopologyId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Delete network topology request failed. error: %v", dbErr)
		return NewErrorResponse(dbErr), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(networkTopologyModel.Name).WithEntityType(NetworkTopologiesAuditEntityType).WithEntityId(networkTopologyModel.ID.String())
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Delete network topology %s request completed successfully", networkTopologyModel.Name)
	return api.Response(http.StatusNoContent, nil), nil
}

// Helper functions

func networkTopologyDaoToApi(networkTopology models.NetworkTopology) api.NetworkTopology {
	apiModel := api.NetworkTopology{
		Id:            networkTopology.ID.String(),
		Name:          networkTopology.Name,
		Levels:        networkTopology.Levels,
		Source:        networkTopology.Source,
		ClusterId:     networkTopology.ClusterID.String(),
		TenantId:      networkTopology.TenantID,
		CreatedAt:     networkTopology.CreatedAt,
		Phase:         networkTopology.Phase,
		GroveTopology: adapters.NewGroveDBModelAdapter(networkTopology.Grove).ToAPIModel(),
	}
	if networkTopology.CreatedBy != nil {
		apiModel.CreatedBy = *networkTopology.CreatedBy
	}
	return apiModel
}

func networkTopologyViewDaoToApi(networkTopologyView models.NetworkTopologyView) api.NetworkTopology {
	apiModel := networkTopologyDaoToApi(networkTopologyView.NetworkTopology)
	if len(networkTopologyView.NodePools) > 0 {
		if np := networkTopologyView.GetNodepools(); np != nil {
			apiModel.NodePools = &np
		}
	}
	return apiModel
}

func networkTopologyDeleteValidations(networkTopology models.NetworkTopology) error {
	// Add any specific validation for deletion if needed
	return nil
}
