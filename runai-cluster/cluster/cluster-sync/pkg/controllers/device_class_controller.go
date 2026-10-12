/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	resourcev1 "k8s.io/api/resource/v1"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
)

func AddDeviceClassHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	deviceClassInformer := cachedClient.GetInformer(cachedclient.DeviceClassType)
	if deviceClassInformer == nil {
		// resource.k8s.io (DRA) is not served on this cluster; nothing to track.
		zap.S().Warn("DeviceClass API is not served (DRA not installed); skipping its controller")
		return
	}

	deviceClassHandler := handlers.NewDeviceClassHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(deviceClassHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DeviceClassRL, deviceClassHandler)

	_, err := deviceClassInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			// Unwrap DeletedFinalStateUnknown tombstones so a missed deletion still
			// reaches DeleteFunc, which is where the per-item DELETE is issued.
			switch filteredObject(obj).(type) {
			case *resourcev1.DeviceClass:
				return backendClient.IsRunaiBackendAvailable()
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*resourcev1.DeviceClass)
				genericControllerErrorHandler(deviceClassHandler.HandleDeviceClassAdd(object),
					constants.AddOperation, "deviceClass", object)
			},
			DeleteFunc: func(obj any) {
				object, ok := deletedObject[*resourcev1.DeviceClass](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(deviceClassHandler.HandleDeviceClassDelete(object),
					constants.DeleteOperation, "deviceClass", object)
			},
			UpdateFunc: func(_ any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*resourcev1.DeviceClass)))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to deviceClass informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent deviceClass controller has started")
}
