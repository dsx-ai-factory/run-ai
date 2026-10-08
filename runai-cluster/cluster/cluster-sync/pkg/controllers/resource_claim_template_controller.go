/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"go.uber.org/zap"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// rctResyncKey coalesces every informer event onto a single workqueue entry.
// The tracker accepts only a full-set sync (there is no per-item CRUD surface
// for RCTs - the single POST is intentional, see the handler for the why), so
// keying by object key would just manifest as a burst of redundant resyncs.
var rctResyncKey = client.ObjectKey{Name: "resource-claim-templates"}

func AddResourceClaimTemplateHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	rctInformer := cachedClient.GetInformer(cachedclient.ResourceClaimTemplateType)
	if rctInformer == nil {
		// resource.k8s.io (DRA) is not served on this cluster; nothing to track.
		zap.S().Warn("ResourceClaimTemplate API is not served (DRA not installed); skipping its controller")
		return
	}

	handler := handlers.NewResourceClaimTemplateHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(handler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ResourceClaimTemplateRL, handler)

	_, err := rctInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			// Unwrap DeletedFinalStateUnknown tombstones so a missed deletion still
			// reaches DeleteFunc and triggers the coalesced full-set resync; without
			// this the informer would silently drop the tombstone here and the
			// tracker would keep the stale RCT row until the next event.
			switch filteredObject(obj).(type) {
			case *resourcev1.ResourceClaimTemplate:
				return backendClient.IsRunaiBackendAvailable()
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(_ any) {
				rateLimitedController.AddRateLimited(rctResyncKey)
			},
			UpdateFunc: func(_ any, _ any) {
				rateLimitedController.AddRateLimited(rctResyncKey)
			},
			DeleteFunc: func(_ any) {
				rateLimitedController.AddRateLimited(rctResyncKey)
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to ResourceClaimTemplate informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent ResourceClaimTemplate controller has started")
}
