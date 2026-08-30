package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type BackendInferenceService struct {
	inferenceClient workload.WithUpdateAction
}

func NewBackendInferenceService(inferenceClient workload.WithUpdateAction) *BackendInferenceService {
	return &BackendInferenceService{inferenceClient: inferenceClient}
}

func (bws *BackendInferenceService) DeleteByUuid(ctx context.Context, uuid string) error {
	return bws.inferenceClient.Delete(ctx, uuid)
}

func (bws *BackendInferenceService) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bws.inferenceClient.Submit(ctx, name, projectId, clusterId, spec)
}

func (bws *BackendInferenceService) Update(ctx context.Context, uuid string, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bws.inferenceClient.Update(ctx, uuid, clusterId, spec)
}
