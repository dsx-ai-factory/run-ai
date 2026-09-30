/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

// K8sObjectsSyncUpdates is the cluster-agent-facing endpoint that streams
// k8s-object asset updates to the cluster. Handler is not yet implemented;
// it will mirror CredentialsSyncUpdates / DatasourceSyncUpdates once the
// k8s-objects sync list and the RCT branch of prepareAssetSyncUpdatesResp
// are in place.
func (s *SyncApiService) K8sObjectsSyncUpdates(ctx context.Context, clusterId string, after int64, clusterSyncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), errors.New("K8sObjectsSyncUpdates not implemented")
}
