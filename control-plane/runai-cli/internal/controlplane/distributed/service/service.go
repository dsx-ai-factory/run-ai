package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type BackendDistributedService struct {
	distributedClient workload.WithSuspendableAction
}

func NewBackendDistributedService(distributedClient workload.WithSuspendableAction) *BackendDistributedService {
	return &BackendDistributedService{distributedClient: distributedClient}
}

func (bws *BackendDistributedService) DeleteByUuid(ctx context.Context, uuid string) error {
	return bws.distributedClient.Delete(ctx, uuid)
}

func (bws *BackendDistributedService) SuspendByUuid(ctx context.Context, uuid string) error {
	return bws.distributedClient.Suspend(ctx, uuid)
}

func (bws *BackendDistributedService) ResumeByUuid(ctx context.Context, uuid string) error {
	return bws.distributedClient.Resume(ctx, uuid)
}

func (bws *BackendDistributedService) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bws.distributedClient.Submit(ctx, name, projectId, clusterId, spec)
}
