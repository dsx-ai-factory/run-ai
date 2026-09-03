/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsub

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/services/hard_delete"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	async_events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

var (
	once    sync.Once
	isReady = false
)

type pubSubHandler[T any] func(ctx context.Context, event *T) error

func Init() {
	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Errorf("Failed to initialize pubsub: %v", err)
	} else {
		isReady = true
	}
}

func Register(internalService api.InternalAPIServicer) error {
	if !isReady {
		return nil
	}

	once.Do(func() {
		runaipubsub.SubscribeAsync(getClusterDeleteRequestHandler(internalService))
		runaipubsub.SubscribeAsync(getClusterPendingDeleteHandler())
		runaipubsub.SubscribeAsync(getProjectDeletionHandler(internalService))
		runaipubsub.SubscribeAsync(getTenantDeletionHandler())
	})

	return nil
}

func getClusterDeleteRequestHandler(internalService api.InternalAPIServicer) pubSubHandler[async_events.ClusterDelete] {
	return func(ctx context.Context, event *async_events.ClusterDelete) error {
		response, err := internalService.DeleteScopeAccessRules(ctx, string(api.SCOPETYPE_CLUSTER), event.ClusterId, "", "")
		if err != nil {
			return err
		}

		if response.Code >= http.StatusBadRequest {
			return fmt.Errorf("failed to perform async cluster access rules deletion. status code: %d error: %v", response.Code, response.Body)
		}

		cache.RemoveClusterPendingDeletion(ctx, event.ClusterId)
		return nil
	}
}

func getClusterPendingDeleteHandler() pubSubHandler[async_events.ClusterPendingDelete] {
	return func(ctx context.Context, event *async_events.ClusterPendingDelete) error {
		cache.MarkClusterPendingDeletion(ctx, event.ClusterId)
		return nil
	}
}

func getProjectDeletionHandler(internalService api.InternalAPIServicer) pubSubHandler[async_events.ProjectDeleted] {
	return func(ctx context.Context, event *async_events.ProjectDeleted) error {
		response, err := internalService.DeleteScopeAccessRules(ctx, string(api.SCOPETYPE_PROJECT), event.ProjectId, "", "")
		if err != nil {
			return err
		}

		if response.Code >= http.StatusBadRequest {
			return fmt.Errorf(
				"failed to perform async project %s access rules deletion. project id: %s, cluster id: %s, status code: %d error: %v",
				event.ProjectName, event.ProjectId, event.ClusterId, response.Code, response.Body)
		}

		return nil
	}
}

func getTenantDeletionHandler() pubSubHandler[async_events.TenantDelete] {
	return func(ctx context.Context, event *async_events.TenantDelete) error {
		service := hard_delete.NewHardDeletionService()
		err := service.DeleteTenant(ctx, event.TenantId)
		if err != nil {
			return err
		}
		return nil
	}
}
