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

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// AddClusterInfoHandlerToCache add a ClusterInfo handler to cache client
func AddClusterInfoHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	clusterInfoHandler := handlers.NewClusterInfoHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(clusterInfoHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ClusterInfoRL, clusterInfoHandler)

	configMapInformer := cachedClient.GetInformer(cachedclient.ConfigMapType)
	_, err := configMapInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch configMap := obj.(type) {
			case *corev1.ConfigMap:
				return configMap.Name == constants.ClusterInfoConfigMapName &&
					configMap.Namespace == runaiv1.DefaultNamespace
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(obj.(*corev1.ConfigMap)))
			},
			DeleteFunc: func(obj any) {
				object := obj.(*corev1.ConfigMap)
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
			},
			UpdateFunc: func(oldObj any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*corev1.ConfigMap)))
			},
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to ClusterInfo informer, error: %s", err.Error())
	}

	zap.S().Debug("ClusterInfo handling has started")
}
