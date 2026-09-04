/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	runaiclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddWorkspaceHandlerToCache(backendClient *runaiclient.BackendClient, cachedClient cachedclient.CachedClient,
	workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	interactiveInformer := cachedClient.GetInformer(cachedclient.InteractiveWorkloadType)
	if interactiveInformer == nil {
		zap.S().Warn("InteractiveWorkload API is not served; skipping its controller")
		return nil
	}
	workspaceHandlerDeprecated := handlers.NewWorkspaceHandlerDeprecated(cachedClient, backendClient)

	backendClient.RegisterSyncHandler(workspaceHandlerDeprecated)

	workspaceHandler := handlers.NewWorkspaceHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.InteractiveWorkloadRL, workspaceHandler)
	deprecatedRLController := cachedClient.NewRateLimitedController(cachedclient.InteractiveWorkloadDeprecatedRL, workspaceHandlerDeprecated)

	_, err := interactiveInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object any) bool {
			//TODO: uncomment when removing deprecated
			//return handlers.FilterWorkloads(object.(*v2alpha1.InteractiveWorkload).Spec.WorkloadFields)
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object any) {
				obj := object.(*v2alpha1.InteractiveWorkload)
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					genericControllerErrorHandler(workloadHandler.HandleInteractiveAdd(obj),
						constants.AddOperation, "interactiveworkload", obj)
				}
				genericControllerErrorHandler(workspaceHandlerDeprecated.HandleAdd(obj),
					constants.AddOperation, "interactiveworkload-deprecated", obj)
			},
			DeleteFunc: func(object any) {
				obj, ok := deletedObject[*v2alpha1.InteractiveWorkload](object)
				if !ok {
					return
				}
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
					genericControllerErrorHandler(workloadHandler.HandleRunaiWorkloadDelete(obj),
						constants.DeleteOperation, "interactiveworkload", obj)
				}
				deprecatedRLController.MarkItemDeleted(objectKeyFromObject(obj))
				genericControllerErrorHandler(workspaceHandlerDeprecated.HandleDelete(obj),
					constants.DeleteOperation, "interactiveworkload-deprecated", obj)
			},
			UpdateFunc: func(oldObject any, newObject any) {
				newInteractiveWorkload := newObject.(*v2alpha1.InteractiveWorkload)
				oldInteractiveWorkload := oldObject.(*v2alpha1.InteractiveWorkload)
				if handlers.FilterWorkloads(newInteractiveWorkload.Spec.WorkloadFields) {
					if newInteractiveWorkload.Status.Phase != oldInteractiveWorkload.Status.Phase {
						genericControllerErrorHandler(workloadHandler.HandleInteractiveUpdate(newInteractiveWorkload),
							constants.UpdateOperation, "interactiveworkload", newInteractiveWorkload)
					} else {
						rateLimitedController.AddRateLimited(objectKeyFromObject(newInteractiveWorkload))
					}
				}
				deprecatedRLController.AddRateLimited(objectKeyFromObject(newInteractiveWorkload))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to interactive workload informer, error: %s", err.Error())
	}

	zap.S().Debug("The workspaces controller has started")
	return rateLimitedController
}
