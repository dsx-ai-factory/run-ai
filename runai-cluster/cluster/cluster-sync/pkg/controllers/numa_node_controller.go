/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"go.uber.org/zap"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clientgocache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nrtCRDName = "noderesourcetopologies.topology.node.k8s.io"

var exitProcess = os.Exit

func IsNRTCRDInstalled(ctx context.Context, reader client.Reader) bool {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := reader.Get(ctx, client.ObjectKey{Name: nrtCRDName}, crd); err != nil {
		if !apierrors.IsNotFound(err) {
			zap.S().Errorf("Failed checking whether NodeResourceTopology CRD is installed: %+v", err)
		}
		return false
	}
	return isNRTCRDAndEstablished(crd)
}

func isNRTCRDAndEstablished(crd *apiextensionsv1.CustomResourceDefinition) bool {
	if crd.Name != nrtCRDName {
		return false
	}
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.Established &&
			condition.Status == apiextensionsv1.ConditionTrue {
			return true
		}
	}
	return false
}

// WatchForNRTCRD restarts cluster-sync when an NRT CRD that was absent at
// startup is installed, allowing a fresh process to initialize its NRT informer.
func WatchForNRTCRD(ctx context.Context, manager ctrl.Manager) error {
	informer, err := manager.GetCache().GetInformer(
		ctx,
		&apiextensionsv1.CustomResourceDefinition{},
	)
	if err != nil {
		return err
	}
	_, err = informer.AddEventHandler(clientgocache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
			return ok && isNRTCRDAndEstablished(crd)
		},
		Handler: clientgocache.ResourceEventHandlerFuncs{
			AddFunc: func(any) {
				zap.S().Info("NodeResourceTopology CRD established; restarting cluster-sync")
				exitProcess(0)
			},
		},
	})
	return err
}

func AddNumaNodeHandlerToCache(
	cachedClient cachedclient.CachedClient,
	handler *handlers.NumaNodeHandler,
) bool {
	nrtInformer := cachedClient.GetInformer(cachedclient.NodeResourceTopologyType)
	if nrtInformer == nil {
		zap.S().Warn("NodeResourceTopology is not installed; NUMA detail sync is disabled")
		return false
	}

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.NumaNodeRL, handler)
	_, err := nrtInformer.AddEventHandler(clientgocache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			nrt, ok := obj.(*nrtv1alpha2.NodeResourceTopology)
			if !ok {
				return false
			}
			shouldSync, err := handler.ShouldSyncNumaNode(nrt.Name)
			if err != nil {
				zap.S().Errorf(
					"Failed checking whether NUMA nodes for physical node %s should be synced: %+v",
					nrt.Name,
					err,
				)
				return false
			}
			if !shouldSync {
				zap.S().Debugf(
					"Skipping NUMA nodes for physical node %s because NUMA-aware scheduling is not enabled",
					nrt.Name,
				)
			}
			return shouldSync
		},
		Handler: clientgocache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				nrt := obj.(*nrtv1alpha2.NodeResourceTopology)
				genericControllerErrorHandler(
					handler.HandleAdd(nrt),
					constants.AddOperation,
					constants.NumaNodeResourceName,
					nrt,
				)
			},
			UpdateFunc: func(_ any, newObj any) {
				nrt := newObj.(*nrtv1alpha2.NodeResourceTopology)
				rateLimitedController.AddRateLimited(objectKeyFromObject(nrt))
			},
			DeleteFunc: func(obj any) {
				nrt, ok := deletedObject[*nrtv1alpha2.NodeResourceTopology](obj)
				if !ok {
					return
				}
				rateLimitedController.MarkItemDeleted(objectKeyFromObject(nrt))
				genericControllerErrorHandler(
					handler.HandleDelete(nrt),
					constants.DeleteOperation,
					constants.NumaNodeResourceName,
					nrt,
				)
			},
		},
	})
	if err != nil {
		zap.S().Fatalf("Failed to add event handler to NodeResourceTopology informer: %s", err.Error())
	}

	zap.S().Debug("NodeResourceTopology handling has started")
	return true
}
