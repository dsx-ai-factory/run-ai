package cluster

import (
	"context"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"net/http"
	"strconv"

	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

// SimpleMockClient only supports fixed set of clusters
type SimpleMockClient struct {
	Clusters map[string]*clusterService.DisplayedCluster
}

func (s SimpleMockClient) Client() ClientInterface {
	return s
}

func (s SimpleMockClient) GetClusters(ctx context.Context) ([]clusterService.DisplayedCluster, *clusterService.Error) {
	tenantId := auth.TenantIdOf(ctx)

	var clusters []clusterService.DisplayedCluster
	for _, cluster := range s.Clusters {
		if cluster.TenantId == tenantId {
			clusters = append(clusters, *cluster)
		}
	}
	return clusters, nil
}

func (s SimpleMockClient) GetClusterByUuid(ctx context.Context, id string) (*clusterService.DisplayedCluster, *clusterService.Error) {
	for _, cluster := range s.Clusters {
		if cluster.Uuid == id {
			return cluster, nil
		}
	}
	return nil, clusterService.NewError(http.StatusNotFound, "cluster not found")
}

func (s SimpleMockClient) GetClustersMinimalData(ctx context.Context) (*clusterService.GetClustersMinimalData200Response, error) {
	tenantId := auth.TenantIdOf(ctx)

	var clusters []clusterService.ClustersMinimalData
	for _, cluster := range s.Clusters {
		if cluster.TenantId == tenantId {
			clusters = append(clusters, clusterService.ClustersMinimalData{
				Uuid:     cluster.Uuid,
				Name:     cluster.Name,
				TenantId: cluster.TenantId,
				Version:  cluster.Version,
				Domain:   cluster.Domain,
			})
		}
	}
	return &clusterService.GetClustersMinimalData200Response{
		Clusters: clusters,
	}, nil
}

func (s SimpleMockClient) GetClusterMinimalDataById(ctx context.Context, clusterId string) (*clusterService.ClustersMinimalData, error) {
	for _, cluster := range s.Clusters {
		if cluster.Uuid == clusterId {
			return &clusterService.ClustersMinimalData{
				Uuid:     cluster.Uuid,
				Name:     cluster.Name,
				TenantId: cluster.TenantId,
				Version:  cluster.Version,
				Domain:   cluster.Domain,
			}, nil
		}
	}
	return nil, clusterService.NewError(http.StatusNotFound, "cluster not found")
}

type SimpleMockClientFactory struct {
	SimpleMockClient
}

func (c SimpleMockClientFactory) Client() ClientInterface {
	return &c.SimpleMockClient
}

func NewSimpleMockClient() SimpleMockClient {
	return SimpleMockClient{
		Clusters: make(map[string]*clusterService.DisplayedCluster),
	}
}

func (s SimpleMockClient) WithCluster(clusterId string, clusterInfo clusterService.DisplayedCluster) SimpleMockClient {
	s.Clusters[clusterId] = &clusterInfo
	return s
}

// WithClusters sets all clusters to a high version so they won't cause issues of capabilities
func (s SimpleMockClient) WithClusters(clustersPerTenant map[string][]string) SimpleMockClient {
	for tenant, clusters := range clustersPerTenant {
		tenantId, _ := strconv.Atoi(tenant)
		for _, clusterId := range clusters {
			s.Clusters[clusterId] = &clusterService.DisplayedCluster{
				Uuid:     clusterId,
				TenantId: int32(tenantId),
				Version:  *clusterService.NewNullableString(convutils.Ref(testutils.DefaultClusterVersionForTesting)),
			}
		}
	}
	return s
}
