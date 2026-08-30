package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type BackendTrainingService struct {
	trainingClient workload.WithSuspendableAction
}

func NewBackendTrainingService(trainingClient workload.WithSuspendableAction) *BackendTrainingService {
	return &BackendTrainingService{trainingClient: trainingClient}
}

func (bws *BackendTrainingService) DeleteByUuid(ctx context.Context, uuid string) error {
	return bws.trainingClient.Delete(ctx, uuid)
}

func (bws *BackendTrainingService) SuspendByUuid(ctx context.Context, uuid string) error {
	return bws.trainingClient.Suspend(ctx, uuid)
}

func (bws *BackendTrainingService) ResumeByUuid(ctx context.Context, uuid string) error {
	return bws.trainingClient.Resume(ctx, uuid)
}

func (bws *BackendTrainingService) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bws.trainingClient.Submit(ctx, name, projectId, clusterId, spec)
}
