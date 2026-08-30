package clusters_service

import (
	"context"
	"fmt"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"sync"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	clusterClient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// Client is the interface for the cluster service client
type Client interface {
	GetClustersCount(ctx context.Context, groupBy string) ([]openapi.ClustersCount, *openapi.Error)
}

// client is the implementation of the cluster service client, with caching capabilities
type cacheClient struct {
	client clusterClient.IClusterClient
}

var onceClient sync.Once
var deFactoClient cacheClient

// GetClient returns the cluster service client
func GetClient() Client {
	onceClient.Do(func() {
		err := cache.RegisterStore[[]openapi.ClustersCount](options.Ttl(runai_config.Get[*config.ClusterCache]().Expiration()))
		if err != nil {
			logging.Warnf("Failed to register cache store for type clusters count: %v", err)
		}

		deFactoClient = cacheClient{
			client: clusterClient.NewClusterClient(runai_config.Get[*config.ServicesConfig]().ClusterServiceUrl),
		}
	})
	// holds the real client
	return &deFactoClient
}

func SetMock(mockClient clusterClient.IClusterClient) {
	deFactoClient.client = mockClient
	cache.ClearCache()
}

type getClustersCountKey struct {
	tenantId int32
	groupBy  string
}

func (k *getClustersCountKey) String() string {
	return fmt.Sprintf("%d %s", k.tenantId, k.groupBy)
}

func (c *cacheClient) GetClustersCount(ctx context.Context, groupBy string) ([]openapi.ClustersCount, *openapi.Error) {
	// check if the result is in the cache
	if res, found := getClustersCountFromCache(ctx, groupBy); found {
		return res, nil
	}
	clustersCount, apiErr := c.client.GetClustersCount(ctx, groupBy)
	if apiErr == nil {
		key := getClustersCountKey{
			tenantId: auth.TenantIdOf(ctx),
			groupBy:  groupBy,
		}
		err := cache.Set[[]openapi.ClustersCount](ctx, key.String(), clustersCount)
		if err != nil {
			logging.Warnf("Failed to set clusters count in cache for key %s: %v", key.String(), err)
		}
	}
	return clustersCount, apiErr
}

// the following section relate to caching the result from the cluster service,
// as most of the queries are not expected to change frequently, we can cache the result for a certain period of time

func getClustersCountFromCache(ctx context.Context, groupBy string) ([]openapi.ClustersCount, bool) {
	key := getClustersCountKey{
		tenantId: auth.TenantIdOf(ctx),
		groupBy:  groupBy,
	}
	clustersCount, err := cache.Get[[]openapi.ClustersCount](ctx, key.String())
	if err != nil {
		logging.Warnf("Failed to get clusters count from cache for key %s: %v", key.String(), err)
		return nil, false
	}
	if clustersCount != nil {
		return *clustersCount, true
	}
	return nil, false
}
