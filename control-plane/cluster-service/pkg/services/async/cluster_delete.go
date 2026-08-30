package async

import (
	"context"

	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/utils"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

type ClusterDeleteSubscriber struct {
	clusterDeleter clusterdelete.ClusterDeleter
}

func NewClusterDeleteSubscriber(clusterDeleter clusterdelete.ClusterDeleter) *ClusterDeleteSubscriber {
	return &ClusterDeleteSubscriber{
		clusterDeleter: clusterDeleter,
	}
}

func (c *ClusterDeleteSubscriber) SubscribeClusterDelete() <-chan error {
	return runaipubsub.SubscribeAsync[events.ClusterDelete](c.clusterDeleteSubscribeHandler)
}

func (c *ClusterDeleteSubscriber) clusterDeleteSubscribeHandler(ctx context.Context, msg *events.ClusterDelete) error {
	tenantId, err := utils.ParseTenantId(ctx)
	if err != nil {
		return err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Received cluster delete event: %v", msg)
	return c.clusterDeleter.DeleteClusterFromDB(ctx, msg.ClusterId, tenantId)
}
