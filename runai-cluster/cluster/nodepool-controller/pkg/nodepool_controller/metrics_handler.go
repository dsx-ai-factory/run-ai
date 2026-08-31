/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"sync"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type MetricsHandler struct {
	cachedClient cachedclient.CachedClient
	Mutex        sync.Mutex
}

func InitMetricsWatch(cachedClient cachedclient.CachedClient) (*MetricsHandler, error) {
	metricsHandler := MetricsHandler{
		cachedClient: cachedClient,
		Mutex:        sync.Mutex{},
	}
	return &metricsHandler, nil
}

func (metricsHandler *MetricsHandler) Start(ctx context.Context, stopCh <-chan struct{}) {
	nodeInformer := metricsHandler.cachedClient.GetNodeInformer()
	_, err := nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			metricsHandler.handleNodeChange(nil, object.(*corev1.Node), watch.Added)
		},
		DeleteFunc: func(object interface{}) {
			metricsHandler.handleNodeChange(nil, object.(*corev1.Node), watch.Deleted)
		},
		UpdateFunc: func(oldObject interface{}, newObject interface{}) {
			if !common.FilterNodeUpdatesForNPController(oldObject.(*corev1.Node), newObject.(*corev1.Node)) {
				return
			}
			metricsHandler.handleNodeChange(oldObject.(*corev1.Node), newObject.(*corev1.Node), watch.Modified)
		},
	})
	if err != nil {
		log.Fatal().Msgf("Failed to add event handler to node informer, error: %s", err.Error())
	}

	metricsHandler.cachedClient.StartInformers(ctx, stopCh)
}

func (metricsHandler *MetricsHandler) handleNodeChange(oldNode, newNode *corev1.Node, eventType watch.EventType) {
	log.Debug().Msgf("Got change event %v; node: %s", eventType, newNode.Name)

	metricsHandler.Mutex.Lock()
	defer metricsHandler.Mutex.Unlock()
	switch eventType {
	case watch.Added:
		metricsHandler.setMetricsOnAddUpdate(nil, newNode)
	case watch.Modified:
		metricsHandler.setMetricsOnAddUpdate(oldNode, newNode)
	case watch.Deleted:
		metricsHandler.setMetricsOnDelete(newNode)
	default:
	}
}

func (metricsHandler *MetricsHandler) setMetricsOnAddUpdate(oldEventNode, newEventNode *corev1.Node) {
	nodes := &corev1.NodeList{}
	err := metricsHandler.cachedClient.List(context.Background(), nodes)
	if err != nil {
		log.Fatal().Msgf("error encountered during metrics update - get nodes, error: %v", err)
		return
	}

	nodePools := &v1alpha1.NodePoolList{}
	err = metricsHandler.cachedClient.List(context.Background(), nodePools)
	if err != nil {
		log.Fatal().Msgf("error encountered during metrics update - get nodepools, error: %v", err)
		return
	}

	for i := range nodes.Items {
		node := &nodes.Items[i]
		updateNodeMetrics(node, nodePools)
	}

	removePreviousNodePoolInCaseItWasDeleted(oldEventNode, newEventNode, nodePools)
}

func updateNodeMetrics(node *corev1.Node, nodePools *v1alpha1.NodePoolList) {
	currentNodePoolName := utils.GetNodePoolNameFromLabels(node.Labels)
	nodeName := node.Name
	log.Debug().Msgf("Updating nodepool metric of node <%v> and nodepool <%v> to be 1",
		nodeName, currentNodePoolName)
	metrics.SetNodeNodePool(nodeName, currentNodePoolName)

	for _, otherNodePool := range nodePools.Items {
		if otherNodePool.Name != currentNodePoolName {
			log.Debug().Msgf("Updating nodepool metric of node <%v> and nodepool <%v> to be 0",
				nodeName, otherNodePool.Name)
			metrics.RemoveNodeNodePool(nodeName, otherNodePool.Name)
		}
	}
}

// removePreviousNodePoolInCaseItWasDeleted - in case the old nodepool was deleted, and we missed it...
func removePreviousNodePoolInCaseItWasDeleted(oldEventNode, newEventNode *corev1.Node, nodePools *v1alpha1.NodePoolList) {
	if oldEventNode == nil || newEventNode == nil {
		return
	}

	previousNodePoolNameOfEventNode := utils.GetNodePoolNameFromLabels(oldEventNode.Labels)
	currentNodePoolNameOfEventNode := utils.GetNodePoolNameFromLabels(newEventNode.Labels)
	if previousNodePoolNameOfEventNode == currentNodePoolNameOfEventNode {
		return
	}

	// validate if nodepool doesn't exist in the list
	nodePoolExists := false
	for _, nodePool := range nodePools.Items {
		if nodePool.Name == previousNodePoolNameOfEventNode {
			nodePoolExists = true
			break
		}
	}

	if !nodePoolExists {
		metrics.RemoveNodeNodePool(newEventNode.Name, previousNodePoolNameOfEventNode)
	}
}

func (metricsHandler *MetricsHandler) setMetricsOnDelete(node *corev1.Node) {
	if node == nil {
		log.Error().Msgf("got nil node in setMetricsOnDelete")
		return
	}
	nodeName := node.Name
	nodePools := &v1alpha1.NodePoolList{}
	err := metricsHandler.cachedClient.List(context.Background(), nodePools)
	if err != nil {
		log.Fatal().Msgf("error encountered during metrics update - get nodepools, error: %v", err)
		return
	}

	for _, nodePool := range nodePools.Items {
		log.Debug().Msgf("Updating nodepool metric of node <%v> and nodepool <%v> to be 0",
			nodeName, nodePool.Name)
		metrics.RemoveNodeNodePool(nodeName, nodePool.Name)
	}
}
