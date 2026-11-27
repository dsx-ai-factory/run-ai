/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
)

// WorkspacesApiService is a service that implements the logic for the Workspace service
type WorkspacesApiService struct{}

// NewWorkspacesApiService creates a default api service
func NewWorkspacesApiService() *WorkspacesApiService {
	return &WorkspacesApiService{}
}

func (s *WorkspacesApiService) CountAssets(ctx context.Context, kind string, name string, scope string,
	projectId int32, departmentId string, clusterId string,
) (api.ImplResponse, error) {
	queryParams, queryError := query_types.NewQueryParams("", name, scope, false, projectId, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	_, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	assetKind, kindErr := query_types.LoadKindParameter(kind)
	if kindErr != nil {
		return NewErrorResponse(kindErr), nil
	}

	db, err := dao.CountAssetMeta(ctx, assetKind, queryParams)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result := api.AssetCountResponse{
		Result: 0,
	}

	if countErr := db.Count(&result.Result); countErr != nil {
		dbErr := api.NewError(http.StatusServiceUnavailable, "failed to count assets", nil)
		return NewErrorResponse(dbErr.WithDetails(countErr)), nil
	}

	return api.Response(http.StatusOK, result), nil
}

// ListWorkspaces is a placeholder, currently implemented by the "old" backend, so it should never arrive here
func (s *WorkspacesApiService) ListWorkspaces(context.Context /*clusterIdParam WAIT_FOR_BEING_DELETED old backend*/, string /*nodeIdParam*/, string /*filterParam*/, string,
	/*sortByParam*/ string /*sortDirectionParam*/, string /*fromParam*/, int32 /*limitParam*/, int32,
) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *WorkspacesApiService) UpdateWorkspaceStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return updateWorkloadStatus(ctx, clusterId, update)
}

func (s *WorkspacesApiService) SyncWorkspaceStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	return syncWorkloadStatus(ctx, clusterId, updates)
}

func (s *WorkspacesApiService) DeleteWorkspaceById(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	return deleteWorkload(ctx, clusterId, workloadId)
}
