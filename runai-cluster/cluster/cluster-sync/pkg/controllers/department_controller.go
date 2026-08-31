/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddDepartmentHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	departmentHandler := handlers.NewDepartmentHandler(backendClient, cachedClient)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DepartmentRL, departmentHandler)

	departmentInformer := cachedClient.GetInformer(cachedclient.DepartmentType)
	_, err := departmentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			department := object.(*kaiv1alpha1.Department)
			genericControllerErrorHandler(departmentHandler.HandleAddUpdate(department),
				constants.AddOperation, constants.DepartmentResourceName, department)
		},
		UpdateFunc: func(oldObject any, newObject any) {
			if oldObject == nil {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObject.(*kaiv1alpha1.Department)))
				return
			}
			newDepartment := newObject.(*kaiv1alpha1.Department)
			shouldUpdateProj := shouldUpdateDepartment(oldObject.(*kaiv1alpha1.Department), newDepartment)
			if shouldUpdateProj {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newDepartment))
			}
		},
		DeleteFunc: func(object any) {
			department := object.(*kaiv1alpha1.Department)
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(department))
			departmentHandler.HandleDelete(department)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to department informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent department controller has started")
}

func shouldUpdateDepartment(old, new *kaiv1alpha1.Department) bool {
	// TODO: currently we only check if the updatedAt was changed, we should change it to updateAt and status change once we will have phase (updating/creating/deleting indicates change in the crd)
	oldUpdatedAt := old.Annotations[commonconstants.UpdatedAtAnnotation]
	newUpdatedAt := new.Annotations[commonconstants.UpdatedAtAnnotation]
	return oldUpdatedAt == "" || newUpdatedAt == "" || oldUpdatedAt != newUpdatedAt
}
