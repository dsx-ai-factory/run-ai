package clusters_metadata_cache

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
)

const (
	cacheKeyPattern                 = "%d-clusters-%s"
	clusterHistoriesCacheKeyPattern = "%d-clusters-histories-%s"

	cacheFieldAllClusters = "all"
)

func SetClusterHistoryOrInvalidate(ctx context.Context, tenantId int32, clusterHistory *models.ClusterHistory) error {
	clusterId := clusterHistory.ClusterID.String()
	key := generateClusterHistoriesKey(tenantId, clusterId)
	data, err := cache.Get[models.ClusterHistories](ctx, key)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("SetClusterHistoryOrInvalidate failed to get cluster histories from cache, tenant %d cluster %s, error: %v", tenantId, clusterId, err)
		return err
	}

	if data == nil || len(*data) == 0 {
		return nil
	}

	*data = append(*data, *clusterHistory)
	err = cache.Set[models.ClusterHistories](ctx, key, *data)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("SetClusterHistoryOrInvalidate failed to set cluster histories from cache, tenant %d cluster %s, error: %v", tenantId, clusterId, err)
		err = cache.Del[models.ClusterHistories](ctx, key)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("SetClusterHistoryOrInvalidate failed to delete cluster histories from cache, tenant %d cluster %s, error: %v", tenantId, clusterId, err)
		}
	}

	return err
}

func GetAllClusterHistories(ctx context.Context, tenantId int32, clusterIds ...string) (*[]models.ClusterHistories, error) {
	var jointErr error
	var failedClusterIds []string
	allClustersHistories := make([]models.ClusterHistories, 0)
	for _, clusterId := range clusterIds {
		key := generateClusterHistoriesKey(tenantId, clusterId)
		clusterHistories, err := cache.Get[models.ClusterHistories](ctx, key)
		if err != nil {
			jointErr = errors.Join(jointErr, err)
			failedClusterIds = append(failedClusterIds, clusterId)
		}
		if clusterHistories != nil {
			allClustersHistories = append(allClustersHistories, *clusterHistories)
		}
	}
	if len(failedClusterIds) > 0 {
		return nil, fmt.Errorf("failed to get cluster histories for tenant %d clusters %s: %w", tenantId, strings.Join(failedClusterIds, ","), jointErr)
	}
	return &allClustersHistories, nil
}

func SetAllClusterHistories(ctx context.Context, tenantId int32, clustersHistories map[string]models.ClusterHistories) error {
	var jointErr error
	var failedClusterIds []string
	for clusterId, clusterHistories := range clustersHistories {
		key := generateClusterHistoriesKey(tenantId, clusterId)
		err := cache.Set[models.ClusterHistories](ctx, key, clusterHistories)
		if err != nil {
			jointErr = errors.Join(jointErr, err)
			failedClusterIds = append(failedClusterIds, clusterId)
		}
	}
	if len(failedClusterIds) > 0 {
		return fmt.Errorf("failed to set cluster histories for tenant %d clusters %s: %w", tenantId, strings.Join(failedClusterIds, ","), jointErr)
	}
	return nil
}

func GetOrSetClusterMetadata(ctx context.Context, tenantId int32, clusterId string, getter func() (*models.Cluster, error)) (*models.Cluster, error) {
	key := generateKey(tenantId, clusterId)
	cluster, err := cache.Get[models.Cluster](ctx, key)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetOrSetClusterMetadata failed to get cluster from cache, tenant %d, cluster %s, error: %v", tenantId, clusterId, err)
	}
	if cluster != nil {
		return cluster, nil
	}
	cluster, err = getter()
	if err != nil {
		return nil, err
	}
	if cluster != nil {
		if err := cache.Set[models.Cluster](ctx, key, *cluster); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetOrSetClusterMetadata failed to set cluster to cache, tenant %d, cluster %s, error: %v", tenantId, clusterId, err)
		}
	}
	return cluster, nil
}

func GetOrSetAllClustersMetadata(ctx context.Context, tenantId int32, getter func() (*[]models.Cluster, error)) (*[]models.Cluster, error) {
	key := generateKey(tenantId, cacheFieldAllClusters)
	allClusters, err := cache.Get[[]models.Cluster](ctx, key)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetOrSetAllClustersMetadata failed to get all clusters from cache, tenant %d, error: %v", tenantId, err)
	}
	if allClusters != nil {
		return allClusters, nil
	}
	allClusters, err = getter()
	if err != nil {
		return nil, err
	}
	if allClusters != nil {
		if err := cache.Set[[]models.Cluster](ctx, key, *allClusters); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetOrSetAllClustersMetadata failed to set all clusters to cache, tenant %d, error: %v", tenantId, err)
		}
	}
	return allClusters, nil
}

func SetClusterMetadataOnCache(ctx context.Context, tenantId int32, clusterId string, data *models.Cluster) error {
	if data == nil {
		return nil
	}
	return cache.Set[models.Cluster](ctx, generateKey(tenantId, clusterId), *data)
}

func InvalidateClusterMetadata(ctx context.Context, tenantId int32, clusterId string) error {
	return cache.Del[models.Cluster](ctx, generateKey(tenantId, clusterId))
}

func InvalidateAllClustersMetadata(ctx context.Context, tenantId int32) error {
	return cache.Del[[]models.Cluster](ctx, generateKey(tenantId, cacheFieldAllClusters))
}

func generateKey(tenantId int32, clusterId string) string {
	return fmt.Sprintf(cacheKeyPattern, tenantId, clusterId)
}

func generateClusterHistoriesKey(tenantId int32, clusterId string) string {
	return fmt.Sprintf(clusterHistoriesCacheKeyPattern, tenantId, clusterId)
}
