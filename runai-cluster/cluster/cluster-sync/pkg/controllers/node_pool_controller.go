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
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddNodePoolHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient, nodeHandler *handlers.NodeHandler) {
	nodePoolHandler := handlers.NewNodePoolHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(nodePoolHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.NodepoolRL, nodePoolHandler)

	nodePoolInformer := cachedClient.GetInformer(cachedclient.NodePoolType)
	_, err := nodePoolInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			np := object.(*v1alpha1.NodePool)
			genericControllerErrorHandler(nodePoolHandler.HandleAdd(np), constants.AddOperation, constants.NodePoolResourceName, np)
		},
		DeleteFunc: func(object any) {
			np, ok := deletedObject[*v1alpha1.NodePool](object)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(np))
			genericControllerErrorHandler(nodePoolHandler.HandleDelete(np), constants.DeleteOperation, constants.NodePoolResourceName, np)
		},
		UpdateFunc: func(oldObject any, newObject any) {
			oldNp := oldObject.(*v1alpha1.NodePool)
			newNp := newObject.(*v1alpha1.NodePool)
			if oldNp.Spec.PreferredNetworkTopologyName != newNp.Spec.PreferredNetworkTopologyName {
				go nodeHandler.SyncNodesForNodepool(newNp.Name)
			}
			if handlers.IsNumaEnabled(oldNp) != handlers.IsNumaEnabled(newNp) {
				go nodeHandler.SyncNumaNodesForNodepool(newNp.Name)
			}
			if !filterNodePoolsUpdate(oldNp, newNp) {
				return
			}
			rateLimitedController.AddRateLimited(objectKeyFromObject(newNp))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to node pool informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent node pool controller has started")
}

func filterNodePoolsUpdate(old, new *v1alpha1.NodePool) bool {
	// The GPU network acceleration detection result is a controller-written annotation on
	// the kai.resources CRD rather than a status field, so a change to it is a status
	// change as far as the control-plane is concerned.
	if old.Generation != new.Generation ||
		old.Status.Phase != new.Status.Phase ||
		old.Status.Message != new.Status.Message ||
		old.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetected] !=
			new.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetected] ||
		len(old.Status.Nodes) != len(new.Status.Nodes) {
		return true
	}

	oldStatusNodes := map[string]v1alpha1.NodeInNodePool{}
	for _, oldNodeStatus := range old.Status.Nodes {
		oldStatusNodes[oldNodeStatus.Name] = oldNodeStatus
	}

	for _, newNodeStatus := range new.Status.Nodes {
		if oldNodeStatus, ok := oldStatusNodes[newNodeStatus.Name]; !ok ||
			oldNodeStatus.Status != newNodeStatus.Status {
			return true
		}
	}

	return false
}
