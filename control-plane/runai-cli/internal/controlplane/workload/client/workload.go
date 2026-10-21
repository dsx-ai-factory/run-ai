/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/httputils"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/nim/client"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	nimclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

// Create infra for sorting direction and types
const (
	AscendingStr          = "asc"
	CreatedAtStr          = "createdAt"
	desiredPhaseFieldName = "desiredPhase"
)

type WorkloadType string

type typedWorkloadResponse interface {
	openapi.Training |
		openapi.Inference |
		openapi.Workspace |
		openapi.DistributedWorkload |
		openapi.DistributedInference
}

type BackendWorkloadServiceClient struct {
	apiClient *openapi.APIClient
	urlStr    string
	token     string
}

func NewWorkloadServiceClient(urlStr, token string) (*BackendWorkloadServiceClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &BackendWorkloadServiceClient{
		apiClient: openapi.NewAPIClient(conf),
		urlStr:    urlStr,
		token:     token,
	}, nil
}

func (bsc *BackendWorkloadServiceClient) List(ctx context.Context, limit, offset int32, filters []string, deleted bool) (*openapi.GetWorkloads200Response, error) {
	request := bsc.apiClient.WorkloadsAPI.GetWorkloads(ctx)
	request = request.Limit(limit)
	request = request.Offset(offset)
	request = request.Deleted(deleted)

	if len(filters) != 0 {
		request = request.FilterBy(filters)
	}
	request = request.SortOrder(AscendingStr)
	request = request.SortBy(CreatedAtStr)

	workloadList, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(workloadList, response, err)
}

// ListSlim lists the active workloads with only their identity fields, for looking workloads up by name: the control
// plane reads the workloads table instead of the workloads view, so new workloads are found immediately. Control
// planes that don't have the slim endpoint are served by List.
func (bsc *BackendWorkloadServiceClient) ListSlim(ctx context.Context, limit, offset int32, filters []string) (*openapi.GetWorkloadsSlimResponse, error) {
	request := bsc.apiClient.WorkloadsAPI.GetWorkloadsSlim(ctx)
	request = request.Limit(limit)
	request = request.Offset(offset)

	if len(filters) != 0 {
		request = request.FilterBy(filters)
	}

	workloadList, response, err := request.Execute() //nolint:bodyclose // handle in handler
	if response != nil && response.StatusCode == http.StatusNotFound {
		if response.Body != nil {
			_ = response.Body.Close()
		}
		fullWorkloadList, err := bsc.List(ctx, limit, offset, filters, false)
		if err != nil {
			return nil, err
		}
		return ToSlimResponse(fullWorkloadList), nil
	}
	return httputils.HandleResponse(workloadList, response, err)
}

// ToSlimResponse keeps only the slim fields of a workload list.
func ToSlimResponse(workloadList *openapi.GetWorkloads200Response) *openapi.GetWorkloadsSlimResponse {
	slimList := &openapi.GetWorkloadsSlimResponse{
		Next:      workloadList.Next,
		Workloads: make([]openapi.WorkloadSlim, 0, len(workloadList.Workloads)),
	}
	for _, workload := range workloadList.Workloads {
		slimList.Workloads = append(slimList.Workloads, openapi.WorkloadSlim{
			Id:                   workload.Id,
			Name:                 workload.Name,
			Type:                 openapi.PtrString(workload.Type),
			ProjectName:          workload.ProjectName,
			ClusterId:            workload.ClusterId,
			DistributedFramework: workload.DistributedFramework,
		})
	}
	return slimList
}

func (bsc *BackendWorkloadServiceClient) GetWorkload(ctx context.Context, id string) (*openapi.WorkloadDetailed, error) {
	request := bsc.apiClient.WorkloadsAPI.GetWorkload(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(workload, response, err)
}

func (bsc *BackendWorkloadServiceClient) GetTraining(ctx context.Context, id string) (*openapi.Training, error) {
	request := bsc.apiClient.TrainingsAPI.GetTraining(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return handleTypedWorkloadResponse(workload, response, err)
}

func (bsc *BackendWorkloadServiceClient) GetInference(ctx context.Context, id string) (*openapi.Inference, error) {
	request := bsc.apiClient.InferencesAPI.GetInference(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return handleTypedWorkloadResponse(workload, response, err)
}

func (bsc *BackendWorkloadServiceClient) GetWorkspace(ctx context.Context, id string) (*openapi.Workspace, error) {
	request := bsc.apiClient.WorkspacesAPI.GetWorkspace(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return handleTypedWorkloadResponse(workload, response, err)
}

func (bsc *BackendWorkloadServiceClient) GetDistributed(ctx context.Context, id string) (*openapi.DistributedWorkload, error) {
	request := bsc.apiClient.DistributedAPI.GetDistributed(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return handleTypedWorkloadResponse(workload, response, err)
}

func (bsc *BackendWorkloadServiceClient) GetDistributedInference(ctx context.Context, id string) (*openapi.DistributedInference, error) {
	request := bsc.apiClient.DistributedInferencesAPI.GetDistributedInference(ctx, id)

	workload, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return handleTypedWorkloadResponse(workload, response, err)
}

// V2 compatibility snapshots do not persist desiredPhase, while the V1 typed response
// models require it to be one of the generated enum values. Keep the typed spec usable
// by ignoring only the known empty value; all other response errors remain unchanged.
func handleTypedWorkloadResponse[T typedWorkloadResponse](model *T, response *http.Response, responseErr error) (*T, error) {
	if responseErr == nil || response == nil || !httputils.CheckResponseStatusCodeOK(response) {
		return httputils.HandleResponse(model, response, responseErr)
	}

	var apiErr *openapi.GenericOpenAPIError
	if !errors.As(responseErr, &apiErr) {
		return httputils.HandleResponse(model, response, responseErr)
	}

	responseBody := map[string]json.RawMessage{}
	if err := json.Unmarshal(apiErr.Body(), &responseBody); err != nil {
		return httputils.HandleResponse(model, response, responseErr)
	}

	var desiredPhase string
	desiredPhaseJSON, found := responseBody[desiredPhaseFieldName]
	if !found || json.Unmarshal(desiredPhaseJSON, &desiredPhase) != nil || desiredPhase != "" {
		return httputils.HandleResponse(model, response, responseErr)
	}

	delete(responseBody, desiredPhaseFieldName)
	normalizedBody, err := json.Marshal(responseBody)
	if err != nil {
		return httputils.HandleResponse(model, response, responseErr)
	}

	var recovered T
	if err := json.Unmarshal(normalizedBody, &recovered); err != nil {
		return httputils.HandleResponse(model, response, responseErr)
	}

	if response.Body != nil {
		_ = response.Body.Close()
	}
	return &recovered, nil
}

func (bsc *BackendWorkloadServiceClient) GetNimService(ctx context.Context, id string) (*nimclient.NimServiceResponse, error) {
	nimClient, err := client.NewNimClient(bsc.urlStr, bsc.token)
	if err != nil {
		return nil, err
	}
	return nimClient.Get(ctx, id)
}

func (bsc *BackendWorkloadServiceClient) BatchDelete(ctx context.Context, ids ...string) (*openapi.BatchResponse, error) {
	batch := openapi.NewWorkloadBatch(ids, openapi.ACTION_DELETE)
	request := bsc.apiClient.WorkloadsBatchAPI.BatchWorkloads(ctx)
	request = request.WorkloadBatch(*batch)

	bdr, response, err := request.Execute()
	return httputils.HandleResponse(bdr, response, err)
}
