/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

// AddApplicationHandlerToCache add a AiApplication handler to cache client
func AddApplicationHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	applicationHandler := handlers.NewApplicationHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(applicationHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ApplicationRL, applicationHandler)

	applicationInformer := cachedClient.GetInformer(cachedclient.ApplicationType)
	_, err := applicationInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			object := obj.(*v1alpha1.Application)
			genericControllerErrorHandler(applicationHandler.HandleAddUpdate(object), constants.AddOperation, "Application", object)
		},
		DeleteFunc: func(obj any) {
			object, ok := deletedObject[*v1alpha1.Application](obj)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
			genericControllerErrorHandler(applicationHandler.HandleDelete(object), constants.DeleteOperation, "Application", object)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*v1alpha1.Application)))
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to application informer, error: %s", err.Error())
	}

	zap.S().Debug("application handling has started")
}
