/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	cluster_delete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type TenantDeleteSubscriber struct {
	clusterDeleter cluster_delete.ClusterDeleter
}

func NewTenantDeleteSubscriber(clusterDeleter cluster_delete.ClusterDeleter) *TenantDeleteSubscriber {
	return &TenantDeleteSubscriber{
		clusterDeleter: clusterDeleter,
	}
}

func (t *TenantDeleteSubscriber) SubscribeTenantDelete() <-chan error {
	return runaipubsub.SubscribeAsync[events.TenantDelete](t.TenantDeleteSubscribeHandler)
}

func (t *TenantDeleteSubscriber) TenantDeleteSubscribeHandler(ctx context.Context, event *events.TenantDelete) error {
	intTenantId, err := strconv.Atoi(event.TenantId)
	if err != nil {
		return fmt.Errorf("failed to convert tenant id to int: %v", err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Received a request for Tenant:[%d] deletion", intTenantId)

	return t.clusterDeleter.DeleteTenantClusters(ctx, intTenantId)
}
