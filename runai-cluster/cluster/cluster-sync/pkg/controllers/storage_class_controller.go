/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	storagev1 "k8s.io/api/storage/v1"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
)

func AddStorageClassHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	storageClassHandler := handlers.NewStorageClassHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(storageClassHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.StorageClassRL, storageClassHandler)

	storageClassInformer := cachedClient.GetInformer(cachedclient.StorageClassType)
	_, err := storageClassInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch obj.(type) {
			case *storagev1.StorageClass:
				return backendClient.IsRunaiBackendAvailable()
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*storagev1.StorageClass)
				genericControllerErrorHandler(storageClassHandler.HandleStorageClassAdd(object),
					constants.AddOperation, "storageClass", object)
			},
			DeleteFunc: func(obj any) {
				object := obj.(*storagev1.StorageClass)
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(storageClassHandler.HandleStorageClassDelete(object),
					constants.DeleteOperation, "storageClass", object)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*storagev1.StorageClass)))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to storageClass informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent storageClass controller has started")
}
