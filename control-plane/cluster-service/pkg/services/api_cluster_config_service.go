package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	extapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	clusterconfig "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-config"
	status_post_update_handlers "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-config/status-post-update-handlers"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"k8s.io/utils/ptr"
)

type ClusterConfigApiService struct {
	compositeHandler *status_post_update_handlers.CompositeHandler
}

func NewClusterConfigApiService(compositeHandler *status_post_update_handlers.CompositeHandler) api.ClusterConfigAPIServicer {
	return &ClusterConfigApiService{
		compositeHandler: compositeHandler,
	}
}

func (s *ClusterConfigApiService) GetClusterConfig(ctx context.Context, clusterUuid string) (api.ImplResponse, error) {
	clusterConfig, response := s.getClusterConfig(ctx, clusterUuid)
	if response != nil {
		return *response, nil
	}
	return api.Response(http.StatusOK, clusterConfigToClusterConfigResponse(clusterConfig)), nil
}

func (s *ClusterConfigApiService) GetInternalClusterConfig(ctx context.Context, clusterUuid string) (api.ImplResponse, error) {
	clusterConfig, response := s.getClusterConfig(ctx, clusterUuid)
	if response != nil {
		return *response, nil
	}
	return api.Response(http.StatusOK, clusterConfigConfigToGetClusterInternalConfigResponse(clusterConfig)), nil
}

func (s *ClusterConfigApiService) PatchClusterConfig(ctx context.Context, clusterUuid string, patchClusterConfigRequest api.PatchClusterConfigRequest) (api.ImplResponse, error) {
	clusterConfig, response := s.getClusterConfig(ctx, clusterUuid)
	if response != nil {
		return *response, nil
	}
	mergedClusterConfig := clusterconfig.MergeClusterConfig(patchClusterConfigRequest.Config, clusterConfig.Config.UserConfig)
	_, err := dao.SetClusterUserConfig(ctx, clusterConfig.TenantID, clusterUuid, mergedClusterConfig)
	if err != nil {
		baseErr := fmt.Sprintf("Failed to update cluster config, cluster:[%s], tenantId:[%d]", clusterUuid, clusterConfig.TenantID)
		return NewErrorResponse(TranslateDaoError(err, baseErr)), nil
	}
	return api.Response(http.StatusOK, clusterConfigToClusterConfigResponse(clusterConfig)), nil
}

func (s *ClusterConfigApiService) PutClusterConfig(ctx context.Context, clusterUuid string, putClusterConfigRequest api.PutClusterConfigRequest) (api.ImplResponse, error) {
	tenantId, apiErr := extractTenantIdFromCtxAndValidateClusterExistence(ctx, clusterUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterSystemConfig, err := dao.SetClusterUserConfig(ctx, tenantId, clusterUuid, &putClusterConfigRequest.Config)
	if err != nil {
		baseErr := fmt.Sprintf("Failed to update cluster user config, cluster:[%s], tenantId:[%d]", clusterUuid, tenantId)
		return NewErrorResponse(TranslateDaoError(err, baseErr)), nil
	}
	return api.Response(http.StatusOK, clusterConfigToClusterConfigResponse(clusterSystemConfig)), nil
}

func (s *ClusterConfigApiService) PutClusterConfigStatus(ctx context.Context, clusterUuid string, putClusterConfigStatusRequest api.PutClusterConfigStatusRequest) (api.ImplResponse, error) {
	tenantId, apiErr := extractTenantIdFromCtxAndValidateClusterExistence(ctx, clusterUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	clusterStatus := &models.ClusterConfigStatus{
		ActualConfig: putClusterConfigStatusRequest.Status.ActualConfig,
		Conditions:   putClusterConfigStatusRequest.Status.Conditions,
	}
	_, err := dao.SetClusterConfigStatus(ctx, tenantId, clusterUuid, clusterStatus)
	if err != nil {
		baseErr := fmt.Sprintf("Failed to update cluster status, tenantId:[%d], cluster:[%s]", tenantId, clusterUuid)
		return NewErrorResponse(TranslateDaoError(err, baseErr)), nil
	}
	go s.postStatusUpdateHandlersRun(ctx, clusterUuid, tenantId)
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *ClusterConfigApiService) getClusterConfig(ctx context.Context, clusterUuid string) (clusterConfig *models.ClusterConfig, apiResponse *api.ImplResponse) {
	tenantId, apiErr := extractTenantIdFromCtxAndValidateClusterExistence(ctx, clusterUuid)
	if apiErr != nil {
		return nil, ptr.To(NewErrorResponse(apiErr))
	}
	clusterConfig, err := dao.GetClusterConfigByClusterId(ctx, tenantId, clusterUuid)
	if err != nil {
		baseErr := fmt.Sprintf("Failed to get cluster config, cluster:[%s], tenantId:[%d]", clusterUuid, tenantId)
		apiResponse = ptr.To(NewErrorResponse(TranslateDaoError(err, baseErr)))
		return
	}
	return
}

func clusterConfigConfigToGetClusterInternalConfigResponse(icc *models.ClusterConfig) api.GetInternalClusterConfigResponse {
	return api.GetInternalClusterConfigResponse{
		Id:        icc.ID.String(),
		TenantId:  icc.TenantID,
		ClusterId: icc.ClusterID.String(),
		CreatedAt: icc.CreatedAt,
		UpdatedAt: icc.UpdatedAt,
		DeletedAt: func() *time.Time {
			if icc.DeletedAt.Valid {
				return ptr.To(icc.DeletedAt.Time)
			}
			return nil
		}(),
		Config: api.ClusterFullConfigData{
			SystemConfig: icc.Config.SystemConfig,
			UserConfig:   icc.Config.UserConfig,
		},
	}
}

func clusterConfigToClusterConfigResponse(icc *models.ClusterConfig) api.GetPutPatchClusterConfigResponse {
	serviceConfig := runai_config.Get[*config.Service]()
	state, statusMessage := clusterconfig.ResolveClusterState(icc, serviceConfig.MaxWaitClusterRemoteConfigStatusUpdate)
	return api.GetPutPatchClusterConfigResponse{
		Id:        icc.ID.String(),
		TenantId:  icc.TenantID,
		ClusterId: icc.ClusterID.String(),
		CreatedAt: icc.CreatedAt,
		UpdatedAt: icc.UpdatedAt,
		DeletedAt: func() *time.Time {
			if icc.DeletedAt.Valid {
				return ptr.To(icc.DeletedAt.Time)
			}
			return nil
		}(),
		Config: icc.Config.UserConfig,
		Status: api.ClusterConfigStatus{
			State:        state,
			StateMessage: statusMessage,
			ActualConfig: icc.Config.UserConfig,
		},
	}
}

func (s *ClusterConfigApiService) postStatusUpdateHandlersRun(ctx context.Context, clusterUuid string, tenantId int32) {
	copyCtxMap := runaictx.ToMap(ctx)
	ctx = runaictx.FromMap(context.Background(), copyCtxMap)
	clusterConfig, err := dao.GetClusterConfigByClusterId(ctx, tenantId, clusterUuid)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Could not find get cluster config, tenantId:<%d>, clusterUId:<%s>. err: %v", tenantId, clusterUuid, err)
		return
	}
	cluster, apiErr := dao.GetClusterByIdAndTenantId(ctx, clusterUuid, tenantId, extapi.VerbosityMetadata)
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cluster not found in the db probably already deleted, tenantId: <%d>, clusterUId:<%s>", tenantId, clusterUuid)
			return
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Failed to get cluster, tenantId: <%d>, clusterUId:<%s>. err: %v", tenantId, clusterUuid, err)
		return
	}
	s.compositeHandler.HandleClusterConfigStatusChange(ctx, cluster, clusterConfig)
}
