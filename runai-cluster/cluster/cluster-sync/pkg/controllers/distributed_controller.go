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
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddDistributedHandlerToCache(backendClient *runaiclient.BackendClient, cachedClient cachedclient.CachedClient,
	workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	distributedInformer := cachedClient.GetInformer(cachedclient.DistributedWorkloadType)
	if distributedInformer == nil {
		zap.S().Warn("DistributedWorkload API is not served; skipping its controller")
		return nil
	}
	distributedHandlerDeprecated := handlers.NewDistributedHandlerDeprecated(cachedClient, backendClient)

	backendClient.RegisterSyncHandler(distributedHandlerDeprecated)

	distributedHandler := handlers.NewDistributedHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DistributedWorkloadRL, distributedHandler)
	deprecatedRLController := cachedClient.NewRateLimitedController(cachedclient.DistributedWorkloadDeprecatedRL, distributedHandlerDeprecated)

	_, err := distributedInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object any) bool {
			//TODO: uncomment when removing deprecated
			//handlers.FilterWorkloads(object.(*diwv2.DistributedWorkload).Spec.WorkloadFields)
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object any) {
				obj := object.(*diwv2.DistributedWorkload)
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					genericControllerErrorHandler(workloadHandler.HandleDistributedAdd(obj),
						constants.AddOperation, constants.DistributedWorkloadResourceName, obj)
				}
				genericControllerErrorHandler(distributedHandlerDeprecated.HandleAdd(obj),
					constants.AddOperation, "distributedworkload-deprecated", obj)
			},
			DeleteFunc: func(object any) {
				obj := object.(*diwv2.DistributedWorkload)
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
					genericControllerErrorHandler(workloadHandler.HandleRunaiWorkloadDelete(object.(client.Object)),
						constants.DeleteOperation, constants.DistributedWorkloadResourceName, object.(client.Object))
				}
				deprecatedRLController.MarkItemDeleted(objectKeyFromObject(obj))
				genericControllerErrorHandler(distributedHandlerDeprecated.HandleDelete(obj),
					constants.DeleteOperation, "distributedworkload-deprecated", obj)
			},
			UpdateFunc: func(oldObject any, newObject any) {
				newDistributedWorkload := newObject.(*diwv2.DistributedWorkload)
				if handlers.FilterWorkloads(newDistributedWorkload.Spec.WorkloadFields) {
					if newDistributedWorkload.Status.Phase != oldObject.(*diwv2.DistributedWorkload).Status.Phase {
						genericControllerErrorHandler(workloadHandler.HandleDistributedUpdate(newDistributedWorkload),
							constants.UpdateOperation, constants.DistributedWorkloadResourceName, newDistributedWorkload)
					} else {
						rateLimitedController.AddRateLimited(objectKeyFromObject(newDistributedWorkload))
					}
				}
				deprecatedRLController.AddRateLimited(objectKeyFromObject(newDistributedWorkload))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to distributed informer, error: %s", err.Error())
	}

	zap.S().Debug("The distributed controller has started")
	return rateLimitedController
}
