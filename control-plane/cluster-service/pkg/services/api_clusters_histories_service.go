/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ClustersHistoriesApiService struct{}

func NewClustersHistoriesApiService() *ClustersHistoriesApiService {
	return &ClustersHistoriesApiService{}
}

func (s *ClustersHistoriesApiService) GetClustersHistories(ctx context.Context, clusterIds []string) (api.ImplResponse, error) {
	tenantId, apiErr := ParseTenantId(ctx)
	if apiErr != nil {
		return api.Response(http.StatusBadRequest, nil), nil
	}
	clusters, err := dao.GetClustersHistoriesByIds(ctx, tenantId, clusterIds)
	if err != nil {
		return api.Response(http.StatusInternalServerError, api.NewError(http.StatusInternalServerError, "Failed to get clusters history", err)), nil
	}

	return api.Response(http.StatusOK, s.clusterHistoryDaoToApi(clusters)), nil
}

func (s *ClustersHistoriesApiService) clusterHistoryDaoToApi(clustersHistories []*models.ClusterHistory) []api.ClusterHistory {
	parsed := []api.ClusterHistory{}

	for _, clusterHistory := range clustersHistories {
		if len(clusterHistory.Version) == 0 {
			logging.Debugf("empty version, skipped")
			continue
		}

		parsed = append(parsed, api.ClusterHistory{
			ClusterId:       clusterHistory.ClusterID.String(),
			Version:         clusterHistory.Version,
			UpdateTimestamp: clusterHistory.UpdateTimestamp.Unix(),
		})
	}

	return parsed
}
