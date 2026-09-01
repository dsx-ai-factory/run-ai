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
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddTrainingHandlerToCache(backendClient *runaiclient.BackendClient, cachedClient cachedclient.CachedClient,
	workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	trainingInformer := cachedClient.GetInformer(cachedclient.TrainingWorkloadType)
	if trainingInformer == nil {
		zap.S().Warn("TrainingWorkload API is not served; skipping its controller")
		return nil
	}
	trainingHandlerDeprecated := handlers.NewTrainingHandlerDeprecated(cachedClient, backendClient)

	backendClient.RegisterSyncHandler(trainingHandlerDeprecated)

	trainingHandler := handlers.NewTrainingHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.TrainingWorkloadRL, trainingHandler)
	deprecatedRLController := cachedClient.NewRateLimitedController(cachedclient.TrainingWorkloadDeprecatedRL, trainingHandlerDeprecated)

	_, err := trainingInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object any) bool {
			//TODO: uncomment when removing deprecated
			//handlers.FilterWorkloads(object.(*v2alpha1.TrainingWorkload).Spec.WorkloadFields)
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object any) {
				obj := object.(*v2alpha1.TrainingWorkload)
				if handlers.FilterWorkloads(object.(*v2alpha1.TrainingWorkload).Spec.WorkloadFields) {
					genericControllerErrorHandler(workloadHandler.HandleTrainAdd(obj),
						constants.AddOperation, "trainingworkload", obj)
				}
				genericControllerErrorHandler(trainingHandlerDeprecated.HandleAdd(obj),
					constants.AddOperation, "trainingworkload-deprecated", obj)
			},
			DeleteFunc: func(object any) {
				obj, ok := deletedObject[*v2alpha1.TrainingWorkload](object)
				if !ok {
					return
				}
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
					genericControllerErrorHandler(workloadHandler.HandleRunaiWorkloadDelete(obj),
						constants.DeleteOperation, "trainingworkload", obj)
				}
				deprecatedRLController.MarkItemDeleted(objectKeyFromObject(obj))
				genericControllerErrorHandler(trainingHandlerDeprecated.HandleDelete(obj),
					constants.DeleteOperation, "trainingworkload-deprecated", obj)
			},
			UpdateFunc: func(oldObject any, newObject any) {
				newTrainingWorkload := newObject.(*v2alpha1.TrainingWorkload)
				if handlers.FilterWorkloads(newTrainingWorkload.Spec.WorkloadFields) {
					if newTrainingWorkload.Status.Phase != oldObject.(*v2alpha1.TrainingWorkload).Status.Phase {
						genericControllerErrorHandler(workloadHandler.HandleTrainUpdate(newTrainingWorkload), constants.UpdateOperation, "trainingworkload", newTrainingWorkload)
					} else {
						rateLimitedController.AddRateLimited(objectKeyFromObject(newTrainingWorkload))
					}
				}
				deprecatedRLController.AddRateLimited(objectKeyFromObject(newTrainingWorkload))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to training informer, error: %s", err.Error())
	}

	zap.S().Debug("The training controller has started")
	return rateLimitedController
}
