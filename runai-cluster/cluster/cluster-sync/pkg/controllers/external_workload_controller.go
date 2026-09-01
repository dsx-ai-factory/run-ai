/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	runaiclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddExternalWorkloadHandlerToCache(backendClient *runaiclient.BackendClient,
	cachedClient cachedclient.CachedClient, workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	backendClient.RegisterSyncHandler(workloadHandler)

	externalWorkloadHandler := handlers.NewExternalWorkloadHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ExternalWorkloadRL, externalWorkloadHandler)

	externalWorkloadInformer := cachedClient.GetInformer(cachedclient.ExternalWorkloadType)
	_, err := externalWorkloadInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			ew := obj.(*v1alpha1.ExternalWorkload)
			genericControllerErrorHandler(workloadHandler.HandleAdd(ew), constants.AddOperation, "externalworkload", ew)
		},
		DeleteFunc: func(obj any) {
			ew, ok := deletedObject[*v1alpha1.ExternalWorkload](obj)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(ew))
			genericControllerErrorHandler(workloadHandler.HandleDelete(ew), constants.DeleteOperation, "externalworkload", ew)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			newEW := newObj.(*v1alpha1.ExternalWorkload)
			if oldObj.(*v1alpha1.ExternalWorkload).Status.ObservedStatus.Phase != newEW.Status.ObservedStatus.Phase {
				genericControllerErrorHandler(workloadHandler.HandleUpdate(newEW), constants.UpdateOperation, "externalworkload", newEW)
			} else {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*v1alpha1.ExternalWorkload)))
			}
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to external workload informer, error: %s", err.Error())
	}

	zap.S().Debug("ExternalWorkload handling has started")
	return rateLimitedController
}
