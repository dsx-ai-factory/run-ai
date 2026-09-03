/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cleaners

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	cluster_delete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type ClusterCleaner struct {
	clusterDeleter                  cluster_delete.ClusterDeleter
	maxWaitForClusterGracefulDelete time.Duration
	runDuration                     time.Duration
}

func NewClusterCleaner(
	clusterDeleter cluster_delete.ClusterDeleter,
	maxWaitForClusterGracefulDelete time.Duration,
	runDuration time.Duration,
) *ClusterCleaner {
	return &ClusterCleaner{
		clusterDeleter:                  clusterDeleter,
		maxWaitForClusterGracefulDelete: maxWaitForClusterGracefulDelete,
		runDuration:                     runDuration,
	}
}

func (c *ClusterCleaner) Init(ctx context.Context) {
	err := c.handlePendingDeletionClusters(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to handle pending deletion clusters: %v", err)
	}
	go c.pendingDeletionWorker(ctx)
}

func (c *ClusterCleaner) pendingDeletionWorker(ctx context.Context) {
	ticker := time.NewTicker(c.runDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.handlePendingDeletionClusters(ctx)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to handle pending deletion clusters: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *ClusterCleaner) handlePendingDeletionClusters(ctx context.Context) error {
	clusters, error := dao.GetPendingDeletionClusters(ctx)
	if error != nil {
		return fmt.Errorf("failed to get pending deletion clusters: %s", error.Message)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Found %d pending deletion clusters", len(clusters))
	if len(clusters) == 0 {
		return nil
	}
	timeForForceDelete := time.Now().Add(-c.maxWaitForClusterGracefulDelete)
	for _, cluster := range clusters {
		if cluster.DeleteRequestedAt.Before(timeForForceDelete) {
			if err := c.clusterDeleter.ForceDeleteCluster(ctx, &cluster); err != nil {
				return fmt.Errorf("failed to force delete cluster %s: %s", cluster.ID.String(), err.Message)
			}
		}
	}
	return nil
}
