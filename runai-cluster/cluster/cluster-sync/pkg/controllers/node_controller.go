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
	constants2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	clientgocache "k8s.io/client-go/tools/cache"
)

func AddNodeHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient,
	nodeHandler *handlers.NodeHandler) {
	backendClient.RegisterSyncHandler(nodeHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.NodeRL, nodeHandler)

	nodeInformer := cachedClient.GetInformer(cachedclient.NodeType)
	_, err := nodeInformer.AddEventHandler(clientgocache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			node := obj.(*corev1.Node)
			if node.Labels != nil && node.Labels[v1alpha1.NodePoolNameLabel] == constants2.ExcludedNodepoolName {
				zap.S().Debugf("Node %s is excluded from the cluster", node.Name)
				return
			}
			genericControllerErrorHandler(nodeHandler.HandleAdd(node), constants.AddOperation, constants.NodeResourceName, node)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			newNode := newObj.(*corev1.Node)
			if newNode.Labels != nil && newNode.Labels[v1alpha1.NodePoolNameLabel] == constants2.ExcludedNodepoolName {
				oldNode := oldObj.(*corev1.Node)
				if oldNode.Labels != nil && oldNode.Labels[v1alpha1.NodePoolNameLabel] != constants2.ExcludedNodepoolName {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(oldNode))
					genericControllerErrorHandler(nodeHandler.HandleDelete(newNode), constants.DeleteOperation, constants.NodeResourceName, newNode)
				}
				zap.S().Debugf("Node %s is excluded from the cluster", newNode.Name)
				return

			}
			oldNode := oldObj.(*corev1.Node)
			if nodeHandler.NodeChanged(oldNode, newNode) {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*corev1.Node)))
			} else {
				zap.S().Debug("Node %s update skipped, node relevant fields were not changed", newNode.Name)
			}
		},
		DeleteFunc: func(obj any) {
			node, ok := deletedObject[*corev1.Node](obj)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(node))
			genericControllerErrorHandler(nodeHandler.HandleDelete(node), constants.DeleteOperation, constants.NodeResourceName, node)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to node informer, error: %s", err.Error())
	}

	zap.S().Debug("Node handling has started")
}
