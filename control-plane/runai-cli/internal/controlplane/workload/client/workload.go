package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim/client"

	nimclient "github.com/run-ai/runai/runai-clients/nim-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
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
