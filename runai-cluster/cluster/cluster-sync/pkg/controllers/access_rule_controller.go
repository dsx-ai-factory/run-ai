/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"k8s.io/client-go/tools/cache"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRuleController struct {
	_accessRuleHandler *handlers.AccessRuleHandler //nolint:unused
	_watchClient       client.WithWatch            //nolint:unused
}

func AddAccessRuleHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	accessRuleHandler := handlers.NewAccessRuleHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(accessRuleHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.AccessRuleRL, accessRuleHandler)

	accessRuleInformer := cachedClient.GetInformer(cachedclient.AccessRuleType)
	_, err := accessRuleInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			obj := object.(*v1alpha1.AccessRule)
			genericControllerErrorHandler(accessRuleHandler.HandleAdd(obj),
				constants.AddOperation, constants.AccessRuleResourceName, obj)
		},
		DeleteFunc: func(object any) {
			obj, ok := deletedObject[*v1alpha1.AccessRule](object)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
			genericControllerErrorHandler(accessRuleHandler.HandleDelete(obj),
				constants.DeleteOperation, constants.AccessRuleResourceName, obj)
		},
		UpdateFunc: func(oldObject any, newObject any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObject.(*v1alpha1.AccessRule)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to access rule informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent access rule controller has started")
}
