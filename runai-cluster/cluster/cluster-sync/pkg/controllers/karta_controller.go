/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"go.uber.org/zap"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"

	"k8s.io/client-go/tools/cache"
)

// AddKartaHandlerToCache add a Karta handler to cache client
func AddKartaHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	kartaHandler := handlers.NewKartaHandler(backendClient, cachedClient)

	backendClient.RegisterSyncHandler(kartaHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.KartaRL, kartaHandler)

	kartaInformer := cachedClient.GetInformer(cachedclient.KartaType)
	_, err := kartaInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			object := obj.(*kartav1alpha1.Karta)
			genericControllerErrorHandler(kartaHandler.HandleAddUpdate(object), constants.AddOperation, constants.KartaResourceName, object)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			newKarta := newObj.(*kartav1alpha1.Karta)
			oldKarta := oldObj.(*kartav1alpha1.Karta)
			if isStatusChanged(oldKarta, newKarta) {
				genericControllerErrorHandler(kartaHandler.HandleAddUpdate(newKarta), constants.UpdateOperation, constants.KartaResourceName, newKarta)
			} else {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newKarta))
			}
		},
		DeleteFunc: func(obj any) {
			object, ok := deletedObject[*kartav1alpha1.Karta](obj)
			if !ok {
				return
			}
			genericControllerErrorHandler(kartaHandler.HandleDelete(object), constants.DeleteOperation, constants.KartaResourceName, object)
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to Karta informer, error: %s", err.Error())
	}

	zap.S().Debug("Karta handling has started")
}

func isStatusChanged(oldKarta, newKarta *kartav1alpha1.Karta) bool {
	oldPhase, _ := parsers.ParseBackendKartaConditions(oldKarta)
	newPhase, _ := parsers.ParseBackendKartaConditions(newKarta)
	return oldPhase != newPhase
}
