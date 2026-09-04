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
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// AddConfigMapHandlerToCache add a ConfigMap handler to cache client
func AddConfigMapHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	configMapHandler := handlers.NewConfigMapHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(configMapHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ConfigMapRL, configMapHandler)

	configMapInformer := cachedClient.GetInformer(cachedclient.ConfigMapType)
	_, err := configMapInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch configMap := obj.(type) {
			case *corev1.ConfigMap:
				return backendClient.IsRunaiBackendAvailable() &&
					util.IsUserExposedConfigMapForRunai(configMap) &&
					isRelevantResourceScope(cachedClient, &configMap.ObjectMeta)
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*corev1.ConfigMap)
				genericControllerErrorHandler(configMapHandler.HandleConfigMapAdd(object), constants.AddOperation, "ConfigMap", object)
			},
			DeleteFunc: func(obj any) {
				object, ok := deletedObject[*corev1.ConfigMap](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(configMapHandler.HandleConfigMapDelete(object), constants.DeleteOperation, "ConfigMap", object)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*corev1.ConfigMap)))
			},
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to ConfigMap informer, error: %s", err.Error())
	}

	zap.S().Debug("ConfigMap handling has started")
}
