package services_crud

import (
	"context"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
)

type DistributedService struct {
	Conf                 *config.ServerConfig
	SyncConf             *config.WorkloadsSyncConfig
	tenantSettingsClient clients.TenantSettingsClient
}

func NewDistributedService(tenantSettingsClient clients.TenantSettingsClient) api.DistributedAPIServicer {
	return &DistributedService{
		Conf:                 runai_config.Get[*config.ServerConfig](),
		SyncConf:             runai_config.Get[*config.WorkloadsSyncConfig](),
		tenantSettingsClient: tenantSettingsClient,
	}
}

func (s *DistributedService) CreateDistributed(ctx context.Context, request api.DistributedCreationRequest) (api.ImplResponse, error) {
	return CreateDistributed(ctx, s.tenantSettingsClient, request)
}

func (s *DistributedService) DeleteDistributed(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	res, err := ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_DELETED)
	if res.Code >= 200 && res.Code < 300 {
		res.Body = nil // When deleting a single workload, we don't return a response body for 2xx status codes.
	}
	return res, err
}

func (s *DistributedService) GetDistributed(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return GetWorkloadById(ctx, Distributed, workloadId)
}

func (s *DistributedService) ResumeDistributed(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_RUNNING)
}

func (s *DistributedService) SuspendDistributed(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_STOPPED)
}

func (s *DistributedService) ListDistributedUpdates(ctx context.Context, afterUpdate int64, clusterId string) (api.ImplResponse, error) {
	return ListUpdates(ctx, Distributed, afterUpdate, clusterId, s.SyncConf, nil)
}

func (s *DistributedService) RequestDistributedUpdates(ctx context.Context, afterUpdate int64, clusterId string, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return ListUpdates(ctx, Distributed, afterUpdate, clusterId, s.SyncConf, &syncReport)
}
