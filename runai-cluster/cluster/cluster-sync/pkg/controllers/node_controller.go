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
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddNodeHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient,
	nodeHandler *handlers.NodeHandler) {
	backendClient.RegisterSyncHandler(nodeHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.NodeRL, nodeHandler)

	addGpuFractioningVerdictWatch(cachedClient, nodeHandler)
	addNvidiaResourceSliceWatch(cachedClient, rateLimitedController)

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
				if oldNode.Labels[v1alpha1.NodePoolNameLabel] != constants2.ExcludedNodepoolName {
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

// The fractioning verdict lives in the cluster-info ConfigMap, not on the Node objects whose
// conditions are derived from it, so nothing in the Node informer fires when it flips.
func addGpuFractioningVerdictWatch(cachedClient cachedclient.CachedClient, nodeHandler *handlers.NodeHandler) {
	configMapInformer := cachedClient.GetInformer(cachedclient.ConfigMapType)
	_, err := configMapInformer.AddEventHandler(clientgocache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			configMap, ok := filteredObject(obj).(*corev1.ConfigMap)
			return ok && configMap.Name == constants.ClusterInfoConfigMapName &&
				configMap.Namespace == runaiv1.DefaultNamespace
		},
		Handler: clientgocache.ResourceEventHandlerDetailedFuncs{
			AddFunc: func(obj any, isInInitialList bool) {
				nodeHandler.UpdateGpuFractioningSupport(obj.(*corev1.ConfigMap), isInInitialList)
			},
			UpdateFunc: func(_, newObj any) {
				nodeHandler.UpdateGpuFractioningSupport(newObj.(*corev1.ConfigMap), false)
			},
			DeleteFunc: func(obj any) {
				if _, ok := deletedObject[*corev1.ConfigMap](obj); !ok {
					return
				}
				nodeHandler.UpdateGpuFractioningSupport(nil, false)
			},
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to watch %s for GPU fractioning changes, error: %s",
			constants.ClusterInfoConfigMapName, err.Error())
	}
}

// A node's GPU info can come from the NVIDIA DRA driver's ResourceSlices when its GFD labels are missing. The slices
// are published after the Node object is stable, so nothing in the Node informer fires when they appear or change.
func addNvidiaResourceSliceWatch(cachedClient cachedclient.CachedClient,
	nodeController *cachedclient.RateLimitedController) {
	resourceSliceInformer := cachedClient.GetInformer(cachedclient.ResourceSliceType)
	if resourceSliceInformer == nil {
		zap.S().Warn("ResourceSlice API is not served (DRA not installed); node GPU info is read from labels only")
		return
	}

	resyncNode := func(obj any) {
		resourceSlice, ok := filteredObject(obj).(*resourcev1.ResourceSlice)
		if !ok || resourceSlice.Spec.Driver != resources.NvidiaDraDriverName || resourceSlice.Spec.NodeName == nil ||
			*resourceSlice.Spec.NodeName == "" {
			return
		}
		nodeController.AddRateLimited(client.ObjectKey{Name: *resourceSlice.Spec.NodeName})
	}

	_, err := resourceSliceInformer.AddEventHandler(clientgocache.ResourceEventHandlerFuncs{
		AddFunc:    resyncNode,
		UpdateFunc: func(_ any, newObj any) { resyncNode(newObj) },
		DeleteFunc: resyncNode,
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to resource slice informer, error: %s", err.Error())
	}
}
