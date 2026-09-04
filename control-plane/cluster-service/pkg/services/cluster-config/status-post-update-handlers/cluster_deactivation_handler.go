/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_post_update_handlers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	clusterdelete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type ClusterDeactivationHandler struct {
	clusterDeleter clusterdelete.ClusterDeleter
}

func newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter clusterdelete.ClusterDeleter) Handler {
	return &ClusterDeactivationHandler{
		clusterDeleter: clusterDeleter,
	}
}

func (c ClusterDeactivationHandler) HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig) {
	if !clusterConfig.Status.ActualConfig.SystemConfig.Deactivated {
		return
	}
	err := c.clusterDeleter.ForceDeleteCluster(ctx, cluster)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Failed to initiate cluster force delete, tenantId: <%d>, clusterUId:<%s>. err: %v", cluster.TenantId, cluster.ID.String(), err)
	}
}
