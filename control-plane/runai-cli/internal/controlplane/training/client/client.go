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

type BackendTrainingClient struct {
	apiClient    *openapi.APIClient
	nodesService *nodes.BackendNodeService
}

func NewTrainingClient(urlStr, token string) (*BackendTrainingClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}

	nodesService, err := nodes.CreateBackendNodeService(urlStr, token)
	if err != nil {
		return nil, err
	}

	return &BackendTrainingClient{apiClient: openapi.NewAPIClient(conf), nodesService: nodesService}, nil
}

func (bwc *BackendTrainingClient) Delete(ctx context.Context, uuid string) error {
	_, response, err := bwc.apiClient.TrainingsAPI.DeleteTraining(ctx, uuid).Execute()
	return handleDeleteResponse(response, err)
}

func (bwc *BackendTrainingClient) Suspend(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.apiClient.TrainingsAPI.SuspendTraining(ctx, uuid).Execute())
}

func (bwc *BackendTrainingClient) Resume(ctx context.Context, uuid string) error {
	return handleSuspendResponse(bwc.apiClient.TrainingsAPI.ResumeTraining(ctx, uuid).Execute())
}

func (bwc *BackendTrainingClient) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	request := bwc.apiClient.TrainingsAPI.CreateTraining(ctx)
	training := *openapi.NewTrainingCreationRequest(name, projectId, clusterId)
	wsSpec, err := spec.GenerateTrainingSpec(ctx, clusterId, bwc.nodesService)
	if err != nil {
		return "", err
	}
	training.SetSpec(*wsSpec)
	training.SetUseGivenNameAsPrefix(spec.HasPrefixName())
	if spec.GetTemplateId() != "" {
		training.SetTemplateId(spec.GetTemplateId())
	}
	if assets := spec.GetAssetSubmissionIds(); assets != nil {
		training.SetAssets(*assets)
	}
	if spec.IsDryRun() {
		var byteSpec []byte
		byteSpec, err = training.MarshalJSON()
		return string(byteSpec), err
	}
	request = request.TrainingCreationRequest(training)

	var trainingName string
	trainingName, err = handleSubmitResponse(request.Execute())
	return trainingName, err
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

func handleSubmitResponse(training *openapi.Training, response *http.Response, err error) (string, error) {
	if err != nil {
		return "", clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return training.Name, nil
}
