package metrics_exporter

import (
	"context"

	api "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

//go:generate mockery --name ClustersClienter
type ClustersClienter interface {
	GetClusterIds(ctx context.Context, tenantId int32) ([]string, error)
}

type clustersClient struct {
	client *api.APIClient
}

func newClustersClient(clustersServiceUrl string) *clustersClient {
	return &clustersClient{
		client: api.NewAPIClient(&api.Configuration{
			Servers: api.ServerConfigurations{{URL: clustersServiceUrl}},
		}),
	}
}

func (c *clustersClient) GetClusterIds(ctx context.Context, tenantId int32) ([]string, error) {
	clusters, _, err := c.client.ClustersAPI.GetInternalClusters(ctx).Execute()
	if err != nil {
		return nil, err
	}

	clusterIds := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		clusterIds = append(clusterIds, cluster.Uuid)
	}

	return clusterIds, nil
}
