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
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddDatavolumeHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	datavolumeHandler := handlers.NewDatavolumeHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(datavolumeHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DatavolumeRL, datavolumeHandler)

	datavolumeInformer := cachedClient.GetInformer(cachedclient.DatavolumeType)
	_, err := datavolumeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			obj := object.(*v1alpha1.DataVolume)
			genericControllerErrorHandler(datavolumeHandler.HandleAddUpdate(obj),
				constants.AddOperation, constants.DataVolumeResourceName, obj)
		},
		DeleteFunc: func(object any) {
			obj, ok := deletedObject[*v1alpha1.DataVolume](object)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
			genericControllerErrorHandler(datavolumeHandler.HandleDelete(obj),
				constants.DeleteOperation, constants.DataVolumeResourceName, obj)
		},
		UpdateFunc: func(oldObject any, newObject any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObject.(*v1alpha1.DataVolume)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to datavolume informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent datavolume controller has started")
}
