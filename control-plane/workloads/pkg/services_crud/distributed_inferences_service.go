package services_crud

import (
	"context"
	"fmt"
	"net/http"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
)

const DistributedInferenceResourceType = "distributedInferences"

type DistributedInferenceService struct {
	Conf                 *config.ServerConfig
	SyncConf             *config.WorkloadsSyncConfig
	tenantSettingsClient clients.TenantSettingsClient
	WorkloadsDao         *dao.WorkloadsDao
}

func NewDistributedInferenceService(client clients.TenantSettingsClient) api.DistributedInferencesAPIServicer {
	conf := runai_config.Get[*config.ServerConfig]()
	return &DistributedInferenceService{
		Conf:                 conf,
		SyncConf:             runai_config.Get[*config.WorkloadsSyncConfig](),
		tenantSettingsClient: client,
		WorkloadsDao:         dao.NewWorkloadsDao(conf),
	}
}

func (s *DistributedInferenceService) GetInferenceAncestors(ctx context.Context, resourceId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	ancestors, err := authorization_utils.GetAncestorsCache(ctx, DistributedInferenceResourceType, resourceId)
	if err == nil {
		return api.Response(http.StatusOK, ancestors), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find inference or revision <%s> ancestors in cache. err: %v, continuing", resourceId, err)

	workload, apiErr := services_integration.GetWorkloadDetails(ctx, resourceId)
	if apiErr != nil && apiErr.Code != http.StatusNotFound {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	if apiErr == nil {
		ancestors = &authorization_utils.Ancestors{
			TenantId:     fmt.Sprintf("%d", workload.TenantID),
			ClusterUuid:  ptr.To(workload.ClusterID.String()),
			DepartmentId: ptr.To(workload.DepartmentID),
			ProjectId:    ptr.To(workload.ProjectID),
		}
	} else {
		revision, dbErr := dao.DbGetRevision(ctx, tenantId, resourceId)
		if dbErr != nil {
			return api.Response(int(dbErr.Code), dbErr.Message), nil
		}

		ancestors = &authorization_utils.Ancestors{
			TenantId:     fmt.Sprintf("%d", revision.TenantID),
			ClusterUuid:  ptr.To(revision.ClusterID.String()),
			DepartmentId: ptr.To(revision.DepartmentID),
			ProjectId:    ptr.To(revision.ProjectID),
		}
	}

	authorization_utils.SetAncestorsCache(ctx, DistributedInferenceResourceType, resourceId, *ancestors)

	return api.Response(http.StatusOK, *ancestors), nil
}

func (s *DistributedInferenceService) CreateDistributedInference(ctx context.Context, request api.DistributedInferenceCreationRequest) (api.ImplResponse, error) {
	return CreateDistributedInference(ctx, s.tenantSettingsClient, request)
}

func (s *DistributedInferenceService) ListDistributedInferencesUpdates(ctx context.Context, afterUpdate int64, clusterId string) (api.ImplResponse, error) {
	return ListUpdates(ctx, DistributedInference, afterUpdate, clusterId, s.SyncConf, nil)
}

func (s *DistributedInferenceService) RequestDistributedInferencesUpdates(ctx context.Context, afterUpdate int64, clusterId string, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return ListUpdates(ctx, DistributedInference, afterUpdate, clusterId, s.SyncConf, &syncReport)
}

func (s *DistributedInferenceService) GetDistributedInference(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	return GetWorkloadById(ctx, DistributedInference, workloadId)
}

func (s *DistributedInferenceService) DeleteDistributedInference(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	res, err := ChangeDesiredStates(ctx, []string{workloadId}, api.WORKLOADDESIREDPHASE_DELETED)
	if res.Code >= 200 && res.Code < 300 {
		res.Body = nil // When deleting a single workload, we do not return a response body for 2xx status codes.
	}
	return res, err
}

func (s *DistributedInferenceService) UpdateDistributedInferenceSpec(ctx context.Context, workloadId string, request api.UpdateRequest) (api.ImplResponse, error) {
	return UpdateDistributedInferenceSpec(ctx, workloadId, s.tenantSettingsClient, request)
}
