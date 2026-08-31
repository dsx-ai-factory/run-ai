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
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	scope2 "github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"

	"github.com/cenkalti/backoff/v4"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	workloadsdb "github.com/run-ai/runai/control-plane/assets-service/pkg/database/workloads"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var (
	ReplicaTypeMaster = convutils.Ref(string(workloadsService.REPLICATYPE_MASTER))
	ReplicaTypeWorker = convutils.Ref(string(workloadsService.REPLICATYPE_WORKER))
)

// WorkloadApiService is a service that implements the logic for the workloads service
type WorkloadApiService struct{}

// RUNAI_MODEL and RUNAI_MODEL_NAME are environment variables that are set to the model name in the assets contained in the workload.
var (
	MODEL_ASSET_ENV_VAR_RUNAI_MODEL      = "RUNAI_MODEL"
	MODEL_ASSET_ENV_VAR_RUNAI_MODEL_NAME = "RUNAI_MODEL_NAME"
)

func NewWorkloadApiService() *WorkloadApiService {
	return &WorkloadApiService{}
}

// ValidateWorkloadRequest validates the request's assets and returns a list of the contained assets and their
// related credential and imposed assets
func ValidateWorkloadRequest(ctx context.Context, workloadType string, containedAssetIds []string,
	specificEnv *api.SpecificRunParams,
) ([]dbmodel.AssetsContainedInWorkload, map[string]dbmodel.CredentialsAssetSpecs, *api.Error) {
	var credentialsFromRunParams []string
	if specificEnv != nil {
		credentialsFromRunParams = dao.GetCredentialIdsFromEnvironmentVariables(specificEnv.EnvironmentVariables)
	}

	workloadAssets, credentialsAssets, dbError := dao.ListWorkloadAssets(containedAssetIds, credentialsFromRunParams, runaidb.DB().WithContext(ctx))
	if dbError != nil {
		return nil, nil, dbError
	}

	if specificEnv != nil {
		validationError := validateSpecificEnv(workloadType, workloadAssets, specificEnv)
		if validationError != nil {
			return nil, nil, validationError
		}
	}

	if workloadType != auth.WorkloadTypeTraining && specificEnv != nil {
		if specificEnv.Parallelism != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, "parallelism is only supported for standard training workloads", nil)
		}
		if specificEnv.Completions != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, "completions is only supported for standard training workloads", nil)
		}
	}

	return workloadAssets, credentialsAssets, nil
}

func GetWorkloadById(ctx context.Context, workloadType string, id string) (*dbmodel.GenericWorkload, *api.Error) {
	genericWorkload, getError := workloadsdb.GetWorkloadById(ctx, workloadType, id)
	if getError != nil {
		return nil, getError
	}

	workloadTypeAsAssetKind := auth.WorkloadTypeToAssetKind[workloadType]

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfWorkload(genericWorkload.Meta.ProjectId), string(workloadTypeAsAssetKind), backend.ACTION_READ, false)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	return genericWorkload, nil
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

func ScopeInfoOfWorkload(projectId int32) *scope.ScopeInfo {
	// workloads are always project-scoped
	return scope.ScopeInfoOf(scope.Project, &projectId, nil, nil, nil)
}

// notifyWorkloadService notifies the workloads service on the new workload
// pass creation metadata by value to avoid sharing memory between go routine
func notifyWorkloadService(ctx context.Context, notification workloadsService.WorkloadCreationFields) {
	var err *workloadsService.Error
	exponentialBackoff := backoff.NewExponentialBackOff()
	sc := runai_config.Get[*config.ServerConfig]()
	exponentialBackoff.MaxInterval = time.Duration(sc.RetryMaxIntervalSec) * time.Second
	exponentialBackoff.MaxElapsedTime = time.Duration(sc.RetryMaxIntervalSec) * time.Second
	ticker := backoff.NewTicker(exponentialBackoff)
	for range ticker.C {
		_, err = workloadsServiceClient.Get().Client().InternalCreateWorkload(ctx, &notification)
		if err == nil {
			return
		}
		// Retry only if the error is 5XX
		if err.Code < 500 {
			return
		}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to notify workloads service: workload %s workload type %s error %v", notification.Name,
		notification.Type, err)
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
