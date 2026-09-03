/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"strings"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddPodHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient,
	podHandler *handlers.PodHandler, nodeHandler *handlers.NodeHandler) *cachedclient.RateLimitedController {

	backendClient.RegisterSyncHandler(podHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)

	podInformer := cachedClient.GetInformer(cachedclient.PodType)
	_, err := podInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			switch pod := obj.(type) {
			case *corev1.Pod:
				zap.S().Debug("Pod added in k8s")
				if backendClient.IsRunaiBackendAvailable() && pod.Spec.SchedulerName == constants.RunaiSchedulerName {
					return true
				} else if filterPodForNodeUpdateNeeded(pod) {
					return true
				} else {
					return false
				}
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				pod := obj.(*corev1.Pod)
				if filterPodForNodeUpdateNeeded(pod) {
					updateNode(pod, cachedClient, nodeHandler)
					return
				}
				if pod.Spec.SchedulerName != constants.RunaiSchedulerName {
					return
				}

				genericControllerErrorHandler(podHandler.HandlePodAdd(obj.(*corev1.Pod)), constants.AddOperation, constants.PodResourceName, obj.(*corev1.Pod))
			},
			DeleteFunc: func(obj any) {
				pod, ok := deletedObject[*corev1.Pod](obj)
				if !ok {
					return
				}
				if filterPodForNodeUpdateNeeded(pod) {
					updateNode(pod, cachedClient, nodeHandler)
					return
				}
				if pod.Spec.SchedulerName != constants.RunaiSchedulerName {
					return
				}

				rateLimitedController.MarkItemDeleted(objectKeyFromObject(pod))
				genericControllerErrorHandler(podHandler.HandlePodDelete(pod), constants.DeleteOperation, constants.PodResourceName, pod)
			},
			UpdateFunc: func(oldObj any, newObj any) {
				newPod := newObj.(*corev1.Pod)
				if filterPodForNodeUpdateNeeded(newPod) {
					updateNode(newPod, cachedClient, nodeHandler)
					return
				}
				if newPod.Spec.SchedulerName != constants.RunaiSchedulerName {
					return
				}

				oldPhase := parsers.GetPodRunaiK8sPhase(oldObj.(*corev1.Pod))
				newPhase := parsers.GetPodRunaiK8sPhase(newPod)

				if oldPhase != newPhase {
					genericControllerErrorHandler(podHandler.HandlePodUpdate(newPod), constants.UpdateOperation, constants.PodResourceName, newPod)
				} else {
					rateLimitedController.AddRateLimited(objectKeyFromObject(newPod))
				}
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to pod informer, error: %s", err.Error())
	}

	zap.S().Debug("Pod handling has started")
	return rateLimitedController
}

func filterPodForNodeUpdateNeeded(pod *corev1.Pod) bool {
	return pod.Spec.NodeName != "" && strings.HasPrefix(pod.Name, "nvidia")
}

func updateNode(newPod *corev1.Pod, cachedReader client.Reader, nodeHandler *handlers.NodeHandler) {
	nodeToUpdate := &corev1.Node{}
	err := cachedReader.Get(context.Background(), types.NamespacedName{
		Name: newPod.Spec.NodeName,
	}, nodeToUpdate)
	if err != nil {
		zap.S().Errorf("Failed getting Node <%s>, error: %s", newPod.Spec.NodeName, err.Error())
		return
	}
	zap.S().Infof("Pod <%s/%s> is an nvidia pod - updating Node <%s>", newPod.Namespace, newPod.Name, newPod.Spec.NodeName)
	go genericControllerErrorHandler(nodeHandler.HandleUpdate(nodeToUpdate), constants.UpdateOperation, constants.NodeResourceName, nodeToUpdate)
}
