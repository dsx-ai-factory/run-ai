/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"go.uber.org/zap"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func AddSecretHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	secretHandler := handlers.NewSecretHandler(cachedClient, backendClient)
	backendClient.RegisterSyncHandler(secretHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.SecretRL, secretHandler)

	secretInformer := cachedClient.GetInformer(cachedclient.SecretType)
	_, err := secretInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch secret := obj.(type) {
			case *corev1.Secret:
				return backendClient.IsRunaiBackendAvailable() &&
					(util.IsUserExposedCredentialsForRunai(secret) || util.IsAssetRelatedResource(&secret.ObjectMeta)) &&
					(isRelevantResourceScope(cachedClient, &secret.ObjectMeta) || isUserCredential(secret.ObjectMeta))
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				object := obj.(*corev1.Secret)
				genericControllerErrorHandler(secretHandler.HandleSecretAdd(object), constants.AddOperation, "secret", object)
			},
			DeleteFunc: func(obj any) {
				object, ok := deletedObject[*corev1.Secret](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
				genericControllerErrorHandler(secretHandler.HandleSecretDelete(object), constants.DeleteOperation, "secret", object)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*corev1.Secret)))
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to secret informer, error: %s", err.Error())
	}

	zap.S().Debug("Secret handling has started")
}
