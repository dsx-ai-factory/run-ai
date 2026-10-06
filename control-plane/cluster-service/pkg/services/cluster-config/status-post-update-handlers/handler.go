/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_post_update_handlers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"

	clusterdelete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
)

type Handler interface {
	HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig)
}

type CompositeHandler struct {
	handlers []Handler
}

func NewCompositeHandler(clusterDeleter clusterdelete.ClusterDeleter) *CompositeHandler {
	deactivationHandler := newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter)
	handlers := []Handler{
		deactivationHandler,
	}
	return &CompositeHandler{
		handlers: handlers,
	}
}

func (c CompositeHandler) HandleClusterConfigStatusChange(ctx context.Context, cluster *models.Cluster, clusterConfig *models.ClusterConfig) {
	for _, handler := range c.handlers {
		handler.HandleClusterConfigStatusChange(ctx, cluster, clusterConfig)
	}
}
