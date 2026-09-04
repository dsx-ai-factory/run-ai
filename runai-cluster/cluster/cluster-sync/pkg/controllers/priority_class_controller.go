/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	schedulingv1 "k8s.io/api/scheduling/v1"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

// AddPriorityClassHandlerToCache add a PriorityClass handler to cache client
func AddPriorityClassHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	priorityClassHandler := handlers.NewPriorityClassHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(priorityClassHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.PriorityClassRL, priorityClassHandler)

	priorityClassInformer := cachedClient.GetInformer(cachedclient.PriorityClassType)
	_, err := priorityClassInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			object := obj.(*schedulingv1.PriorityClass)
			genericControllerErrorHandler(priorityClassHandler.HandleAddUpdate(object), constants.AddOperation, "PriorityClass", object)
		},
		DeleteFunc: func(obj any) {
			object, ok := deletedObject[*schedulingv1.PriorityClass](obj)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
			genericControllerErrorHandler(priorityClassHandler.HandleDelete(object), constants.DeleteOperation, "PriorityClass", object)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*schedulingv1.PriorityClass)))
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to priority class informer, error: %s", err.Error())
	}

	zap.S().Debug("Priority class handling has started")
}
