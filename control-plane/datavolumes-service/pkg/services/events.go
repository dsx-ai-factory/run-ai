package services

import (
	"context"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/dbtables"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func HandleTenantDeleteEvent(ctx context.Context, req *events.TenantDelete) error {
	logging.Infof("Handeling tenant delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, dbtables.ScopeTenant, req.TenantId)
}

func HandleClusterDeleteEvent(ctx context.Context, req *events.ClusterDelete) error {
	logging.Infof("Handeling cluster delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, api.SCOPETYPE_CLUSTER, req.ClusterId)
}

func HandleProjectDeleteEvent(ctx context.Context, req *events.ProjectDeleted) error {
	logging.Infof("Handeling project delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, api.SCOPETYPE_PROJECT, req.ProjectId)
}
