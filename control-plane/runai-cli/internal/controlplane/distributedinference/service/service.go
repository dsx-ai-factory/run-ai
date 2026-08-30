package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type BackendDistributedInferenceService struct {
	distributedInferenceClient workload.WithUpdateAction
}

func NewBackendDistributedInferenceService(distributedInferenceClient workload.WithUpdateAction) *BackendDistributedInferenceService {
	return &BackendDistributedInferenceService{distributedInferenceClient: distributedInferenceClient}
}

func (bds *BackendDistributedInferenceService) DeleteByUuid(ctx context.Context, uuid string) error {
	return bds.distributedInferenceClient.Delete(ctx, uuid)
}

func (bds *BackendDistributedInferenceService) Update(ctx context.Context, uuid, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bds.distributedInferenceClient.Update(ctx, uuid, clusterId, spec)
}

func (bds *BackendDistributedInferenceService) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bds.distributedInferenceClient.Submit(ctx, name, projectId, clusterId, spec)
}
