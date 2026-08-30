package client

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type BackendWorkspaceClient struct {
	apiClient    *openapi.APIClient
	nodesService *nodes.BackendNodeService
}

func NewWorkspaceClient(urlStr, token string) (*BackendWorkspaceClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	nodesService, err := nodes.CreateBackendNodeService(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &BackendWorkspaceClient{apiClient: openapi.NewAPIClient(conf), nodesService: nodesService}, nil
}

func (bwc *BackendWorkspaceClient) Delete(ctx context.Context, uuid string) error {
	_, response, err := bwc.apiClient.WorkspacesAPI.DeleteWorkspace(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bwc *BackendWorkspaceClient) Suspend(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.apiClient.WorkspacesAPI.SuspendWorkspace(ctx, uuid).Execute())
}

func (bwc *BackendWorkspaceClient) Resume(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.apiClient.WorkspacesAPI.ResumeWorkspace(ctx, uuid).Execute())
}

func (bwc *BackendWorkspaceClient) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bwc.apiClient.WorkspacesAPI.CreateWorkspace(ctx)
	workspace := *openapi.NewWorkspaceCreationRequest(name, projectId, clusterId)
	wsSpec, err := spec.GenerateWorkspaceSpec(ctx, clusterId, bwc.nodesService)
	if err != nil {
		return "", err
	}
	workspace.SetSpec(*wsSpec)
	workspace.SetUseGivenNameAsPrefix(spec.HasPrefixName())
	if spec.GetTemplateId() != "" {
		workspace.SetTemplateId(spec.GetTemplateId())
	}
	if assets := spec.GetAssetSubmissionIds(); assets != nil {
		workspace.SetAssets(*assets)
	}
	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = workspace.MarshalJSON()
		return string(byteSpec), err
	}
	request = request.WorkspaceCreationRequest(workspace)

	var workspaceName string
	workspaceName, err = handleSubmitResponse(request.Execute())
	return workspaceName, err
}

func handleDeleteResponse(response *http.Response, err error) error {
	if err != nil {
		return clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		return response.Body.Close()
	}
	return nil
}

func handleSuspendResponse(_ *openapi.HttpResponse, response *http.Response, err error) error {
	if err != nil {
		return clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		return response.Body.Close()
	}
	return nil
}

func handleSubmitResponse(workspace *openapi.Workspace, response *http.Response, err error) (string, error) {
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return workspace.Name, nil
}
