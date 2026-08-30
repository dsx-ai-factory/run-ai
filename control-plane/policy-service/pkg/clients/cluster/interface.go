package cluster

import (
	"context"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

type ClientInterface interface {
	GetClusters(ctx context.Context) ([]clusterService.DisplayedCluster, *clusterService.Error)
	GetClusterByUuid(ctx context.Context, id string) (*clusterService.DisplayedCluster, *clusterService.Error)
	GetClustersMinimalData(ctx context.Context) (*clusterService.GetClustersMinimalData200Response, error)
	GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, error)
}
