package services_crud

import (
	"context"

	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
)

type WorkspacesService struct {
	Conf                 *config.ServerConfig
	SyncConf             *config.WorkloadsSyncConfig
	tenantSettingsClient clients.TenantSettingsClient
}

func NewWorkspacesService(tenantSettingsClient clients.TenantSettingsClient) api.WorkspacesAPIServicer {
	return &WorkspacesService{
		Conf:                 runai_config.Get[*config.ServerConfig](),
		SyncConf:             runai_config.Get[*config.WorkloadsSyncConfig](),
		tenantSettingsClient: tenantSettingsClient,
	}
}

func (s *WorkspacesService) CreateWorkspace(ctx context.Context, request api.WorkspaceCreationRequest) (api.ImplResponse, error) {
	return CreateWorkspace(ctx, s.tenantSettingsClient, request)
}

func (s *WorkspacesService) DeleteWorkspace(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	res, err := ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_DELETED)
	if res.Code >= 200 && res.Code < 300 {
		res.Body = nil // When deleting a single workload, we do not return a response body for 2xx status codes.
	}
	return res, err
}

func (s *WorkspacesService) GetWorkspace(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return GetWorkloadById(ctx, Workspace, workloadId)
}

func (s *WorkspacesService) ResumeWorkspace(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_RUNNING)
}

func (s *WorkspacesService) SuspendWorkspace(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_STOPPED)
}

func (s *WorkspacesService) ListWorkspacesUpdates(ctx context.Context, afterUpdate int64, clusterId string) (api.ImplResponse, error) {
	return ListUpdates(ctx, Workspace, afterUpdate, clusterId, s.SyncConf, nil)
}

func (s *WorkspacesService) RequestWorkspacesUpdates(ctx context.Context, afterUpdate int64, clusterId string, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return ListUpdates(ctx, Workspace, afterUpdate, clusterId, s.SyncConf, &syncReport)
}
