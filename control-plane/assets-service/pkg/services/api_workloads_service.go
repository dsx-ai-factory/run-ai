/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	scope2 "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	workloadsdb "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/workloads"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// WorkloadApiService is a service that implements the logic for the workloads service
type WorkloadApiService struct{}

func NewWorkloadApiService() *WorkloadApiService {
	return &WorkloadApiService{}
}

func (s *WorkloadApiService) AssetNameConflictValidation(ctx context.Context, assetKind string, assetName string, scopeType string, scopeId string) (api.ImplResponse, error) {
	if assetKind == "" {
		return NewErrorResponse(api.NewValidationError("mandatory kind parameter is not provided")), nil
	}
	if assetName == "" {
		return NewErrorResponse(api.NewValidationError("mandatory assetName parameter is not provided")), nil
	}
	if scopeType == "" {
		return NewErrorResponse(api.NewValidationError("mandatory scopeType is not provided")), nil
	}
	if scopeId == "" && api.Scope(scopeType) != api.SCOPE_SYSTEM && api.Scope(scopeType) != api.SCOPE_TENANT && api.Scope(scopeType) != api.SCOPE_CLUSTER {
		return NewErrorResponse(api.NewValidationError("mandatory scopeId is not provided")), nil
	}

	var clusterId *string
	var departmentId *string
	var projectId *int32

	if api.Scope(scopeType) == api.SCOPE_CLUSTER {
		clusterId = &scopeId
	} else if api.Scope(scopeType) == api.SCOPE_DEPARTMENT {
		departmentId = &scopeId
	} else if api.Scope(scopeType) == api.SCOPE_PROJECT {
		intScopeId, err := strconv.Atoi(scopeId)
		if err != nil {
			return NewErrorResponse(api.NewValidationError("invalid scopeId provided")), nil
		}
		projectId = convutils.Ref(int32(intScopeId))
	}

	scopeInfo := scope.ScopeInfoOf(scope.Scope(scopeType), projectId, departmentId, clusterId, nil)
	if rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, scopeInfo, assetKind, backend.ACTION_READ, false); rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	nameValidationQuery, err := scope2.GetNameValidationQuery(ctx, assetName, api.Scope(scopeType), projectId, departmentId, clusterId, nil, api.AssetKind(assetKind), nil)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	isOk, err := dao.AssetNameAvailability(ctx, api.AssetKind(assetKind), nameValidationQuery)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	if !isOk {
		return api.Response(http.StatusConflict, api.NewError(http.StatusConflict, "the given asset name is not available", nil)), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

// the following methods are implementation of the old backend functions that handle updates
// from cluster_sync

func updateWorkloadStatus(ctx context.Context, clusterId string, update api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	err := workloadsdb.UpdateWorkloadCurrentState(ctx, clusterId, update.Id, update.Status)
	if err != nil {
		if err.Code == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("record for update %s does not exist: %v", update.Id, err.Message)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update workload status: %v", err.Message)
		}
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, update), nil
}

func syncWorkloadStatus(ctx context.Context, clusterId string, updates []api.WorkloadStatusUpdate) (api.ImplResponse, error) {
	for _, update := range updates {
		err := workloadsdb.UpdateWorkloadCurrentState(ctx, clusterId, update.Id, update.Status)
		if err != nil {
			if err.Code == http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("record for update %s does not exist status: %v", update.Id, err.Message)
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update workload status: %v", err.Message)
			}
		}
	}

	return api.Response(http.StatusOK, updates), nil
}

func deleteWorkload(ctx context.Context, clusterId string, workloadId string) (api.ImplResponse, error) {
	err := workloadsdb.DeleteWorkloadAndAssetsById(ctx, clusterId, workloadId)
	if err != nil {
		if err.Code == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("record for %s does not exist status: %v", workloadId, err.Message)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete workload status: %v", err.Message)
		}
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}
