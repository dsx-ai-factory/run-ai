package status_post_update_handlers

import (
	"context"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"

	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
)

//go:generate mockery --name Handler
type Handler interface {
	HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig)
}

type CompositeHandler struct {
	handlers []Handler
}

func NewCompositeHandler(clusterDeleter clusterdelete.ClusterDeleter) *CompositeHandler {
	deactivationHandler := newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter)
	handlers := []Handler{
		deactivationHandler,
	}
	return &CompositeHandler{
		handlers: handlers,
	}
}

func (c CompositeHandler) HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig) {
	for _, handler := range c.handlers {
		handler.HandleClusterConfigStatusChange(ctx, cluster, clusterConfig)
	}
}
