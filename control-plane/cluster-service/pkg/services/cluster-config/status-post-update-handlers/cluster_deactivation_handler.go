package status_post_update_handlers

import (
	"context"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type ClusterDeactivationHandler struct {
	clusterDeleter clusterdelete.ClusterDeleter
}

func newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter clusterdelete.ClusterDeleter) Handler {
	return &ClusterDeactivationHandler{
		clusterDeleter: clusterDeleter,
	}
}

func (c ClusterDeactivationHandler) HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig) {
	if !clusterConfig.Status.ActualConfig.SystemConfig.Deactivated {
		return
	}
	err := c.clusterDeleter.ForceDeleteCluster(ctx, cluster)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Failed to initiate cluster force delete, tenantId: <%d>, clusterUId:<%s>. err: %v", cluster.TenantId, cluster.ID.String(), err)
	}
}
