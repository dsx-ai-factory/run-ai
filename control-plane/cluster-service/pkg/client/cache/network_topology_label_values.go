package cache

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// NodeNetworkTopologyLabelValues represents topology labels for a node (nullable array of strings)
type NodeNetworkTopologyLabelValues struct {
	LabelValues []string `json:"labelValues"`
}

type NodeNetworkTopologyLabelValuesCache struct {
	clusterUuid string
	nodeName    string
}

func NewNodeNetworkTopologyLabelValuesCache(clusterUuid, nodeName string) *NodeNetworkTopologyLabelValuesCache {
	return &NodeNetworkTopologyLabelValuesCache{
		clusterUuid: clusterUuid,
		nodeName:    nodeName,
	}
}

func (c *NodeNetworkTopologyLabelValuesCache) getNetworkTopologyLabelValuesCacheKey() string {
	return fmt.Sprintf("%s.%s", c.clusterUuid, c.nodeName)
}

func (c *NodeNetworkTopologyLabelValuesCache) GetCachedNetworkTopologyLabelValues(ctx context.Context) (*NodeNetworkTopologyLabelValues, error) {
	cachedLabels, err := cache.Get[NodeNetworkTopologyLabelValues](ctx, c.getNetworkTopologyLabelValuesCacheKey())
	if err != nil {
		errorMessage := fmt.Sprintf("[NodeNetworkTopologyLabelValuesCache.GetCachedNetworkTopologyLabelValues] failed to get NodeNetworkTopologyLabelValues from cache for clusterUuid <%s> node <%s>. err: %v", c.clusterUuid, c.nodeName, err)
		return nil, fmt.Errorf("could not fetch cache data. %s", errorMessage)
	}

	if cachedLabels == nil {
		missMessage := fmt.Sprintf("Cache miss. Resource clusterUuid:<%s> NodeName:<%s> NodeNetworkTopologyLabelValues are not in cache.", c.clusterUuid, c.nodeName)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("%s", missMessage)
		return nil, nil
	}

	hitMessage := fmt.Sprintf("Cache hit. Resource clusterUuid:<%s> NodeName:<%s> NodeNetworkTopologyLabelValues is in cache.", c.clusterUuid, c.nodeName)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("%s", hitMessage)
	return cachedLabels, nil
}

func (c *NodeNetworkTopologyLabelValuesCache) SetCachedNetworkTopologyLabelValues(ctx context.Context, nodeNetworkTopologyLabelValues NodeNetworkTopologyLabelValues) {
	err := cache.Set[NodeNetworkTopologyLabelValues](ctx, c.getNetworkTopologyLabelValuesCacheKey(), nodeNetworkTopologyLabelValues)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set NodeNetworkTopologyLabelValues data to cache, err: %v", err)
	}
}

func (c *NodeNetworkTopologyLabelValuesCache) DeleteCachedNetworkTopologyLabelValues(ctx context.Context) {
	err := cache.Del[NodeNetworkTopologyLabelValues](ctx, c.getNetworkTopologyLabelValuesCacheKey())
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete NodeNetworkTopologyLabelValues data from cache with key %s, err: %v", c.getNetworkTopologyLabelValuesCacheKey(), err)
	}
}
