/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	clusters_metadata_cache "github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao/cache/clusters/metadata"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func GetPermittedScopesFilter(permittedScopes openapi.PermittedScopes, clusterIdColumnName string) string {
	return authorization_utils.GetPermittedScopesDbFilterWithColumnNames(
		permittedScopes,
		"tenant_id",
		clusterIdColumnName,
		"",
		"",
	)
}

func NodeInfoToNodeDao(tenantId int32, clusterUuid, nodeName string, nodeInfo api.NodeInfo) *models.Node {
	node := &models.Node{
		Common: models.Common{
			CreatedAt: nodeInfo.CreatedAt,
		},
		Status:          nodeInfo.Status,
		Conditions:      nodeInfo.Conditions,
		Taints:          nodeInfo.Taints,
		NodePool:        nodeInfo.NodePool,
		GpuType:         "",
		GpuCount:        0,
		Name:            nodeName,
		ClusterUuid:     clusterUuid,
		TenantId:        tenantId,
		NvLinkCliqueId:  nodeInfo.NvLinkCliqueId,
		NvLinkDomainUid: nodeInfo.NvLinkDomainUid,
	}

	if nodeInfo.Numa != nil {
		node.TopologyManagerPolicy = (*string)(nodeInfo.Numa.TopologyManagerPolicy)
	}

	if nodeInfo.GpuInfo != nil {
		node.GpuType = nodeInfo.GpuInfo.GpuType
		node.GpuCount = nodeInfo.GpuInfo.GpuCount
		node.GpuMemoryBytes = nodeInfo.GpuInfo.GpuMemoryBytes
	}

	if nodeInfo.NetworkTopologyLabels != nil {
		labels, values := models.Node{}.ToNetworkTopologyLabelsAndValues(nodeInfo.NetworkTopologyLabels)
		node.NetworkTopologyLabels = labels
		node.NetworkTopologyLabelValues = values
	}

	return node
}

func invalidateAllClustersMetadataAndLog(ctx context.Context, tenantId int32) {
	err := clusters_metadata_cache.InvalidateAllClustersMetadata(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to invalidate metadata cache for all clusters of tenant <%d>, error: %v",
			tenantId, err)
	}
}

func invalidateClusterMetadataAndLog(ctx context.Context, tenantId int32, clusterId string) {
	err := clusters_metadata_cache.InvalidateClusterMetadata(ctx, tenantId, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to invalidate metadata cache for cluster <%s> of tenant <%d>, error: %v",
			clusterId, tenantId, err)
	}
}

func invalidateAllTenantClustersMetaDataAndLog(ctx context.Context, tenantId int32, clusters *[]models.Cluster) {
	invalidateAllClustersMetadataAndLog(ctx, tenantId)
	if clusters == nil {
		return
	}
	for _, cluster := range *clusters {
		invalidateClusterMetadataAndLog(ctx, tenantId, cluster.ID.String())
	}
}

// applySortingParam orders by sortBy directly rather than through a WithEnricher
// scope: scopes run at query execution, which would push the caller's sort behind
// the "id DESC" tiebreaker added by applyPaginationParam.
func applySortingParam(query orm.DB, sortBy string, sortOrder string) orm.DB {
	if sortBy == "" {
		return query
	}
	return query.Order(querydbenrich.FormatSortClause(sortBy, sortOrder))
}

func applyPaginationParam(query orm.DB, offset int32, limit int32) orm.DB {
	query = query.Order("id DESC")

	if offset != 0 {
		query = query.Offset(int(offset))
	}

	if limit != 0 {
		query = query.Limit(int(limit))
	}
	return query
}
