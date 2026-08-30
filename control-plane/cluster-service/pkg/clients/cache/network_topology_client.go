package cache

import (
	"context"

	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/logging"

	cluster_service_cache "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/cache"
)

func Init() {
	if err := cluster_service_client.InitSharedClusterClientCache(); err != nil {
		logging.Warnf("Failed to register shared cache store: %v", err)
	}
}

// NetworkTopologyLabelValuesCache provides methods to interact with the node network topology label values cache
type NetworkTopologyLabelValuesCache interface {
	// Set sets the network topology label values for a node in the cache
	Set(ctx context.Context, clusterUuid, nodeName string, labelValues []string)
	// Delete deletes the network topology label values for a node from the cache
	Delete(ctx context.Context, clusterUuid, nodeName string)
}

type DefaultNetworkTopologyLabelValuesCache struct{}

func NewNetworkTopologyLabelValuesCache() NetworkTopologyLabelValuesCache {
	return &DefaultNetworkTopologyLabelValuesCache{}
}

func (c *DefaultNetworkTopologyLabelValuesCache) Set(ctx context.Context, clusterUuid, nodeName string, labelValues []string) {
	cacheClient := cluster_service_cache.NewNodeNetworkTopologyLabelValuesCache(clusterUuid, nodeName)
	cacheClient.SetCachedNetworkTopologyLabelValues(ctx, cluster_service_cache.NodeNetworkTopologyLabelValues{
		LabelValues: labelValues,
	})
}

func (c *DefaultNetworkTopologyLabelValuesCache) Delete(ctx context.Context, clusterUuid, nodeName string) {
	cacheClient := cluster_service_cache.NewNodeNetworkTopologyLabelValuesCache(clusterUuid, nodeName)
	cacheClient.DeleteCachedNetworkTopologyLabelValues(ctx)
}
