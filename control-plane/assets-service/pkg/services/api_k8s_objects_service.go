/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
)

// K8sObjectsSyncUpdates is the cluster-agent-facing endpoint that streams
// k8s-object asset updates (resource claim templates today) to the cluster.
// It mirrors CredentialsSyncUpdates / DatasourceSyncUpdates: the shared
// handleAgentSyncUpdates orchestrator loads the deltas past `after`, and
// prepareAssetSyncUpdatesResp fans them into UnifiedAssetSyncResponse via
// the RCT branch. K8s-object specs are not encrypted, so needToDecrypt is
// false.
func (s *SyncApiService) K8sObjectsSyncUpdates(ctx context.Context, clusterId string, after int64, clusterSyncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	unifiedResult, err := s.handleAgentSyncUpdates(ctx, clusterId, after, &clusterSyncReport, dbmodel.K8sObjectsAssetsToSync, false /*no need to decrypt*/)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, unifiedResult.ToK8sObjects()), nil
}
