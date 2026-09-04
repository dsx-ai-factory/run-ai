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

func AddPVCHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	pvcHandler := handlers.NewPVCHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(pvcHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.PVCRL, pvcHandler)

	pvcInformer := cachedClient.GetInformer(cachedclient.PersistentVolumeClaimType)
	_, err := pvcInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch pvc := obj.(type) {
			case *corev1.PersistentVolumeClaim:
				return backendClient.IsRunaiBackendAvailable() &&
					(util.IsUserExposedPvcForRunai(pvc) || util.IsAssetRelatedResource(&pvc.ObjectMeta)) &&
					isRelevantResourceScope(cachedClient, &pvc.ObjectMeta)
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*corev1.PersistentVolumeClaim)
				genericControllerErrorHandler(pvcHandler.HandlePVCAdd(object), constants.AddOperation, "pvc", object)
			},
			DeleteFunc: func(obj any) {
				object, ok := deletedObject[*corev1.PersistentVolumeClaim](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(pvcHandler.HandlePVCDelete(object), constants.DeleteOperation, "pvc", object)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*corev1.PersistentVolumeClaim)))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to pvc informer, error: %s", err.Error())
	}

	_, err = pvcInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch pvc := obj.(type) {
			case *corev1.PersistentVolumeClaim:
				return backendClient.IsRunaiBackendAvailable() &&
					util.IsAssetRelatedResource(&pvc.ObjectMeta)
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*corev1.PersistentVolumeClaim)
				genericControllerErrorHandler(pvcHandler.HandlePvcAssetAdd(object), constants.AddOperation, "pvc", object)
			},
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to pvc informer, error: %s", err.Error())
	}

	zap.S().Debug("PVC handling has started")
}
