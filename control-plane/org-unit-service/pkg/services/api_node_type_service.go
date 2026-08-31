/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type NodeTypeApiService struct{}

func NewNodeTypeApiService() api.NodeTypesAPIServicer {
	return &NodeTypeApiService{}
}

func (s *NodeTypeApiService) CreateNodeType(ctx context.Context, createNodeTypeRequest api.NodeTypeCreateRequest) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	clusterId, err := uuid.Parse(createNodeTypeRequest.ClusterId)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse cluster id"), nil
	}
	subjectId, err := ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}
	nodeType := models.NodeType{
		Name:      createNodeTypeRequest.Name,
		ClusterId: clusterId,
		TenantId:  *tenantId,
	}
	if subjectId != nil {
		nodeType.CreatedBy = *subjectId
	}
	nodeType.CreatedAt = time.Now()

	response, err := dao.CreateNodeType(ctx, &nodeType)
	if err != nil {
		errResponse := HandleDbError(ctx, err, "Failed to create nodeType")
		return errResponse, nil
	}

	nodeTypeResponse := api.NodeType{
		Id:        strconv.FormatInt(response.ID, 10),
		Name:      response.Name,
		ClusterId: response.ClusterId.String(),
	}

	return api.Response(http.StatusCreated, nodeTypeResponse), nil
}

func (s *NodeTypeApiService) GetNodeTypes(ctx context.Context, clusterId string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse cluster id"), nil
	}
	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}
	nodeTypes, err := dao.GetNodeTypes(ctx, *tenantId, clusterUuid, permittedScopes)
	if err != nil {
		errResponse := HandleDbError(ctx, err, "Failed to create node type")
		return errResponse, nil
	}

	apiNodeTypes := make([]api.NodeType, len(*nodeTypes))
	for index, nodeType := range *nodeTypes {
		apiNodeTypes[index] = api.NodeType{
			Id:        strconv.FormatInt(nodeType.ID, 10),
			Name:      nodeType.Name,
			ClusterId: nodeType.ClusterId.String(),
		}
	}

	return api.Response(http.StatusOK, apiNodeTypes), nil
}

func (s *NodeTypeApiService) DeleteNodeType(ctx context.Context, nodeTypeId string) (api.ImplResponse, error) {
	nodeTypeUid, err := strconv.ParseInt(nodeTypeId, 10, 64)
	if err != nil {
		errResponse := HandleError(err, http.StatusBadRequest, "delete node type failed")
		return errResponse, nil
	}

	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	subjectId, err := ParseSubjectId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse subject id"), nil
	}

	err = dao.DeleteNodeType(ctx, nodeTypeUid, *tenantId, *subjectId)
	if err != nil {
		errResponse := HandleDbError(ctx, err, "Failed to delete node type")
		return errResponse, nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *NodeTypeApiService) UpdateNodeType(ctx context.Context, nodeTypeId string, nodeTypeUpdateRequest api.NodeTypeUpdateRequest) (api.ImplResponse, error) {
	nodeTypeUid, err := strconv.ParseInt(nodeTypeId, 10, 64)
	if err != nil {
		errResponse := HandleError(err, http.StatusBadRequest, "delete node type failed")
		return errResponse, nil
	}
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	clusterId, err := uuid.Parse(nodeTypeUpdateRequest.ClusterId)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse cluster id"), nil
	}

	nodeType, err := dao.GetNodeType(ctx, *tenantId, clusterId, nodeTypeUid)
	if err != nil {
		errResponse := HandleDbError(ctx, err, "Failed to update node type")
		return errResponse, nil
	}
	subjectId, err := ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	}

	if subjectId != nil {
		nodeType.UpdatedBy = *subjectId
	}
	nodeType.UpdatedAt = time.Now()
	nodeType.Name = nodeTypeUpdateRequest.Name

	updated, err := dao.UpdateNodeType(ctx, nodeType)
	if err != nil {
		errResponse := HandleDbError(ctx, err, "Failed to update node type")
		return errResponse, nil
	}

	response := api.NodeType{
		Id:        strconv.FormatInt(updated.ID, 10),
		Name:      updated.Name,
		ClusterId: updated.ClusterId.String(),
	}

	return api.Response(http.StatusOK, response), nil
}
