/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"fmt"

	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func pendingDeletionKey(clusterUUID string) string {
	return fmt.Sprintf("cluster-pending-deletion:%s", clusterUUID)
}

// IsClusterPendingDeletion returns true if a ClusterPendingDelete event has been
// received for this cluster UUID and the cache key is present.
//
// The key is written by MarkClusterPendingDeletion (called from the
// ClusterPendingDelete pubsub handler) and removed by RemoveClusterPendingDeletion
// (called from the ClusterDelete pubsub handler).
func IsClusterPendingDeletion(ctx context.Context, clusterUUID string) bool {
	val, err := cache.Get[struct{}](ctx, pendingDeletionKey(clusterUUID))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Warnf("Failed to check pending deletion status for cluster %s: %v", clusterUUID, err)
		return false
	}
	return val != nil
}

// MarkClusterPendingDeletion records that a cluster has delete_requested_at set.
// Called by the ClusterPendingDelete pubsub handler.
func MarkClusterPendingDeletion(ctx context.Context, clusterUUID string) {
	if err := cache.Set[struct{}](ctx, pendingDeletionKey(clusterUUID), struct{}{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Warnf("Failed to mark cluster %s as pending deletion in cache: %v", clusterUUID, err)
	}
}

// RemoveClusterPendingDeletion removes the pending-deletion marker for a cluster.
// Called by the ClusterDelete pubsub handler after access rules are deleted.
func RemoveClusterPendingDeletion(ctx context.Context, clusterUUID string) {
	if err := cache.Del[struct{}](ctx, pendingDeletionKey(clusterUUID)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Warnf("Failed to remove cluster %s pending deletion marker from cache: %v", clusterUUID, err)
	}
}

// InitPendingDeletionCache seeds the cache with all clusters that are pending
// deletion (delete_requested_at set, deleted_at not set). It calls the
// cluster-service internal API via the provided client.
// Intended to be called once at startup so that pre-existing pending-deletion
// clusters are not missed when the authorization service restarts.
func InitPendingDeletionCache(ctx context.Context, clustersClient cluster_service_client.IClusterClient) error {
	return initPendingDeletionCache(ctx, func(ctx context.Context) ([]string, error) {
		return fetchPendingDeletionClusterUUIDs(ctx, clustersClient)
	})
}

func fetchPendingDeletionClusterUUIDs(ctx context.Context, clustersClient cluster_service_client.IClusterClient) ([]string, error) {
	clusters, err := clustersClient.GetInternalClusters(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clusters from cluster-service: %w", err)
	}
	var uuids []string
	for _, c := range clusters {
		if c.HasDeleteRequestedAt() && !c.HasDeletedAt() {
			uuids = append(uuids, c.GetUuid())
		}
	}
	return uuids, nil
}

func initPendingDeletionCache(ctx context.Context, fetchFn func(context.Context) ([]string, error)) error {
	uuids, err := fetchFn(ctx)
	if err != nil {
		return fmt.Errorf("initPendingDeletionCache: %w", err)
	}
	for _, uuid := range uuids {
		MarkClusterPendingDeletion(ctx, uuid)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).
		Infof("InitPendingDeletionCache: seeded %d cluster(s) as pending deletion", len(uuids))
	return nil
}
