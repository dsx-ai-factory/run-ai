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
	dinfwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddDistributedInferenceHandlerToCache(backendClient *runaiclient.BackendClient, cachedClient cachedclient.CachedClient,
	workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	distributedInferenceInformer := cachedClient.GetInformer(cachedclient.DistributedInferenceWorkloadType)
	if distributedInferenceInformer == nil {
		zap.S().Warn("DistributedInferenceWorkload API is not served; skipping its controller")
		return nil
	}
	distributedInferenceHandler := handlers.NewDistributedInferenceHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DistributedInferenceWorkloadRL, distributedInferenceHandler)

	_, err := distributedInferenceInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object any) bool {
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object any) {
				obj := object.(*dinfwv1.DistributedInferenceWorkload)
				if obj.Spec.Usage == constants.RunaiWorkloadRelevantUsage {
					genericControllerErrorHandler(distributedInferenceHandler.HandleAdd(obj),
						constants.AddOperation, constants.DistributedInferenceWorkloadResourceName, obj)
				}
			},
			DeleteFunc: func(object any) {
				obj := object.(*dinfwv1.DistributedInferenceWorkload)
				if obj.Spec.Usage == constants.RunaiWorkloadRelevantUsage {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
					genericControllerErrorHandler(workloadHandler.HandleRunaiWorkloadDelete(object.(client.Object)),
						constants.DeleteOperation, constants.DistributedInferenceWorkloadResourceName, obj)
				}
			},
			UpdateFunc: func(oldObject any, newObject any) {
				newWorkload := newObject.(*dinfwv1.DistributedInferenceWorkload)
				oldWorkload := oldObject.(*dinfwv1.DistributedInferenceWorkload)
				if newWorkload.Spec.Usage == constants.RunaiWorkloadRelevantUsage {
					if newWorkload.Status.Phase != oldWorkload.Status.Phase {
						genericControllerErrorHandler(distributedInferenceHandler.HandleUpdate(newWorkload),
							constants.UpdateOperation, constants.DistributedInferenceWorkloadResourceName, newWorkload)
					} else {
						rateLimitedController.AddRateLimited(objectKeyFromObject(newWorkload))
					}
				}
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to distributed inference informer, error: %s", err.Error())
	}

	zap.S().Debug("The distributed inference controller has started")
	return rateLimitedController
}
