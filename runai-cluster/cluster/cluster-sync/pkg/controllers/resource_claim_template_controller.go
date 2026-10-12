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
	"go.uber.org/zap"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/client-go/tools/cache"
)

// AddResourceClaimTemplateHandlerToCache wires the RCT informer to the per-item
// CRUD + full-set sync pattern, mirroring DeviceClass: Add/Delete fire the
// corresponding per-item HTTP call directly, Update feeds the rate-limited
// controller so bursty mutations collapse into one PUT per object. The periodic
// drift reconciliation runs through RegisterSyncHandler → HandleSyncEvent.
func AddResourceClaimTemplateHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	rctInformer := cachedClient.GetInformer(cachedclient.ResourceClaimTemplateType)
	if rctInformer == nil {
		zap.S().Warn("ResourceClaimTemplate API is not served (DRA not installed); skipping its controller")
		return
	}

	handler := handlers.NewResourceClaimTemplateHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(handler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ResourceClaimTemplateRL, handler)

	_, err := rctInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			// Unwrap DeletedFinalStateUnknown tombstones so a missed deletion still
			// reaches DeleteFunc, which is where the per-item DELETE is issued.
			switch filteredObject(obj).(type) {
			case *resourcev1.ResourceClaimTemplate:
				return backendClient.IsRunaiBackendAvailable()
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*resourcev1.ResourceClaimTemplate)
				genericControllerErrorHandler(handler.HandleResourceClaimTemplateAdd(object),
					constants.AddOperation, "resourceClaimTemplate", object)
			},
			DeleteFunc: func(obj any) {
				object, ok := deletedObject[*resourcev1.ResourceClaimTemplate](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(handler.HandleResourceClaimTemplateDelete(object),
					constants.DeleteOperation, "resourceClaimTemplate", object)
			},
			UpdateFunc: func(_ any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*resourcev1.ResourceClaimTemplate)))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to ResourceClaimTemplate informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent ResourceClaimTemplate controller has started")
}
