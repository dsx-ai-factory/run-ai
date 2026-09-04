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
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddNetworkTopologyHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	networkTopologyHandler := handlers.NewNetworkTopologyHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(networkTopologyHandler)

	networkTopologyInformer := cachedClient.GetInformer(cachedclient.TopologyType)
	_, err := networkTopologyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			np := object.(*kaiv1alpha1.Topology)
			genericControllerErrorHandler(networkTopologyHandler.HandleAdd(np), constants.AddOperation, constants.NetworkTopologyResourceName, np)
		},
		UpdateFunc: func(oldObject any, newObject any) {
			oldNp := oldObject.(*kaiv1alpha1.Topology)
			newNp := newObject.(*kaiv1alpha1.Topology)
			if !groveAnnotationsChanged(oldNp, newNp) {
				return
			}
			genericControllerErrorHandler(networkTopologyHandler.HandleUpdate(newNp), constants.UpdateOperation, constants.NetworkTopologyResourceName, newNp)
		},
		DeleteFunc: func(object any) {
			np, ok := deletedObject[*kaiv1alpha1.Topology](object)
			if !ok {
				return
			}
			genericControllerErrorHandler(networkTopologyHandler.HandleDelete(np), constants.DeleteOperation, constants.NetworkTopologyResourceName, np)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to network topology informer, error: %s", err.Error())
	}

	zap.S().Debug("The network topology controller has started")
}

func groveAnnotationsChanged(oldNp, newNp *kaiv1alpha1.Topology) bool {
	for _, key := range []string{
		commonconstants.GroveTopologyAnnotation,
		commonconstants.GroveTopologyResourceVersionAnnotation,
	} {
		if oldNp.Annotations[key] != newNp.Annotations[key] {
			return true
		}
	}
	return false
}
