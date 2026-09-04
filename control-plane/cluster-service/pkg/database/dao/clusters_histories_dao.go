/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	clusters_metadata_cache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao/cache/clusters/metadata"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func GetClustersHistoriesByIds(ctx context.Context, tenantId int32, ids []string) ([]*models.ClusterHistory, error) {
	clusterHistories := getClustersHistoriesFromCache(ctx, tenantId, ids)
	if clusterHistories != nil {
		return clusterHistories, nil
	}

	return getClusterHistoriesFromDBAndSetToCache(ctx, tenantId, ids)
}

func getClusterHistoriesFromDBAndSetToCache(ctx context.Context, tenantId int32, ids []string) ([]*models.ClusterHistory, error) {
	clustersHistories := make([]*models.ClusterHistory, 0)

	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	query := runaidb.DB().WithContext(ctx).
		Table(models.ClusterHistoryTable).
		Where("cluster_id::text in (?)", ids)

	if !isMgmtTenant {
		query = query.Where("tenant_id = ?", tenantId)
	}

	if err := query.Find(&clustersHistories); err != nil {
		return nil, fmt.Errorf("failed to get clusters for tenant <%v>: %w", ids, err)
	}

	err := clusters_metadata_cache.SetAllClusterHistories(ctx, tenantId, parseToMapByIds(clustersHistories))
	if err != nil {
		logging.Errorf("failed to set cluster histories to cache: %v", err)
	}
	return clustersHistories, nil
}

func getClustersHistoriesFromCache(ctx context.Context, tenantId int32, ids []string) []*models.ClusterHistory {
	histories, err := clusters_metadata_cache.GetAllClusterHistories(ctx, tenantId, ids...)
	if err != nil {
		logging.Errorf("failed to get cluster histories from cache: %v", err)
	} else if histories != nil && len(*histories) == len(ids) {
		return parseResponseFromCacheToArray(histories)
	}
	return nil
}

func InsertClusterHistory(ctx context.Context, cluster *models.Cluster) error {
	if cluster.Version == "" {
		return errors.New("not saved history, empty version")
	}

	exist, err := isRecordForClusterVersionExist(ctx, cluster)
	if err != nil {
		return err
	}

	if !exist {
		h := models.ClusterHistory{
			ClusterID:       cluster.ID,
			Version:         cluster.Version,
			UpdateTimestamp: time.Now(),
			TenantId:        cluster.TenantId,
		}
		tx := runaidb.DB().WithContext(ctx).Begin()
		committed := false
		defer orm.FinalizeTransaction(tx, &committed)
		if _, err := tx.Table(models.ClusterHistoryTable).Create(&h); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		committed = true
		return clusters_metadata_cache.SetClusterHistoryOrInvalidate(ctx, cluster.TenantId, &h)
	}

	return nil
}

func parseToMapByIds(histories []*models.ClusterHistory) map[string]models.ClusterHistories {
	parsed := map[string]models.ClusterHistories{}
	for _, h := range histories {
		if parsed[h.ClusterID.String()] == nil {
			parsed[h.ClusterID.String()] = models.ClusterHistories{}
		}

		parsed[h.ClusterID.String()] = append(parsed[h.ClusterID.String()], *h)
	}
	return parsed
}

func parseResponseFromCacheToArray(histories *[]models.ClusterHistories) []*models.ClusterHistory {
	var toReturn []*models.ClusterHistory
	for _, h := range *histories {
		for _, hh := range h {
			toReturn = append(toReturn, &hh)
		}
	}
	return toReturn
}

func isRecordForClusterVersionExist(ctx context.Context, cluster *models.Cluster) (bool, error) {
	clustersHistories, err := GetClustersHistoriesByIds(ctx, cluster.TenantId, []string{cluster.ID.String()})
	if err != nil {
		return false, err
	}

	for _, ch := range clustersHistories {
		if ch.Version == cluster.Version {
			return true, nil
		}
	}
	return false, nil
}
