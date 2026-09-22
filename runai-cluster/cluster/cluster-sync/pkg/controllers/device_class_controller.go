/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	resourcev1 "k8s.io/api/resource/v1"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
)

// deviceClassesResyncKey is a constant key used to coalesce DeviceClass change events into a
// single full-set re-sync (the tracker only accepts the complete set, so per-object keys add
// no value and would only cause a startup storm of redundant syncs).
var deviceClassesResyncKey = client.ObjectKey{Name: "device-classes"}

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
			switch obj.(type) {
			case *resourcev1.DeviceClass:
				return backendClient.IsRunaiBackendAvailable()
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(_ any) {
				rateLimitedController.AddRateLimited(deviceClassesResyncKey)
			},
			UpdateFunc: func(_ any, _ any) {
				rateLimitedController.AddRateLimited(deviceClassesResyncKey)
			},
			DeleteFunc: func(_ any) {
				rateLimitedController.AddRateLimited(deviceClassesResyncKey)
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to deviceClass informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent deviceClass controller has started")
}
