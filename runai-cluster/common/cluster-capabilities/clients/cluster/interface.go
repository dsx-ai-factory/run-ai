package cluster

import (
	"context"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

type FactoryInterface interface {
	Client() ClientInterface
}

type ClientInterface interface {
	GetClusterByUuid(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterInternal(ctx context.Context, clusterId string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, *clusterService.Error)
}
