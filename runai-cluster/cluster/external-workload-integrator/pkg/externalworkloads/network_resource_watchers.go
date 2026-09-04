/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	commonconf "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// networkResourceGVKs are the GVKs of network resources that should be watched
// for changes to update ExternalWorkload endpoints.
var networkResourceGVKs = []schema.GroupVersionKind{
	{Group: "", Version: "v1", Kind: "Service"},
	{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
	{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"},
	{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GRPCRoute"},
	{Group: "gateway.networking.k8s.io", Version: "v1alpha2", Kind: "TLSRoute"},
	{Group: "inference.networking.k8s.io", Version: "v1", Kind: "InferencePool"},
}
var ocpNetworkResourceGVKs = []schema.GroupVersionKind{
	{Group: "", Version: "v1", Kind: "Service"},
	{Group: "route.openshift.io", Version: "v1", Kind: "Route"},
	{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"},
	{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GRPCRoute"},
	{Group: "gateway.networking.k8s.io", Version: "v1alpha2", Kind: "TLSRoute"},
	{Group: "inference.networking.k8s.io", Version: "v1", Kind: "InferencePool"},
}

// setupNetworkResourceWatchers creates informers for network resources and registers
// event handlers that trigger endpoint recomputation for affected ExternalWorkloads.
// Events are filtered to only process resources in runai-managed namespaces.
func setupNetworkResourceWatchers(ctx context.Context, cachedClient cachedclient.CachedClient, handler *networkstatus.NetworkStatusUpdater) {
	resourceGVKs := networkResourceGVKs
	if commonconf.IsOpenshift() {
		resourceGVKs = ocpNetworkResourceGVKs
	}
	for _, gvk := range resourceGVKs {
		informer, err := cachedClient.SetInformerForKind(ctx, gvk)
		if err != nil {
			if meta.IsNoMatchError(err) {
				log.Info().Msgf("NetworkResourceWatchers: GVK <%s> not available in cluster (CRD not installed), skipping", gvk.String())
				continue
			}
			log.Error().Msgf("NetworkResourceWatchers: failed to get informer for GVK <%s>: %s", gvk.String(), err.Error())
			continue
		}

		capturedGVK := gvk
		queue := handler.TasksQueue()
		_, err = informer.AddEventHandler(clientgocache.FilteringResourceEventHandler{
			FilterFunc: func(obj any) bool {
				clientObj, ok := toClientObject(obj)
				if !ok {
					return false
				}
				return isRunaiManagedNamespace(ctx, cachedClient, clientObj.GetNamespace())
			},
			Handler: clientgocache.ResourceEventHandlerFuncs{
				AddFunc: func(obj any) {
					clientObj, ok := toClientObject(obj)
					if !ok {
						return
					}
					ews := handler.FindAffectedEwsForObject(ctx, clientObj, capturedGVK)
					for i := range ews {
						ew := ews[i]
						key := types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}
						task := utils.NewExternalWorkloadTask(utils.AddOperation, key, capturedGVK).SetObject(ew)
						queue.Add(*task)
					}
				},
				UpdateFunc: func(_, newObj any) {
					clientObj, ok := toClientObject(newObj)
					if !ok {
						return
					}
					ews := handler.FindAffectedEwsForObject(ctx, clientObj, capturedGVK)
					for i := range ews {
						ew := ews[i]
						key := types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}
						task := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, capturedGVK).SetObject(ew)
						queue.Add(*task)
					}
				},
				DeleteFunc: func(obj any) {
					clientObj, ok := toClientObject(obj)
					if !ok {
						return
					}
					ews := handler.FindAffectedEwsForObject(ctx, clientObj, capturedGVK)
					for i := range ews {
						ew := ews[i]
						key := types.NamespacedName{Name: ew.Name, Namespace: ew.Namespace}
						task := utils.NewExternalWorkloadTask(utils.DeleteOperation, key, capturedGVK).SetObject(ew)
						queue.Add(*task)
					}
				},
			},
		})
		if err != nil {
			log.Error().Msgf("NetworkResourceWatchers: failed to add event handler for GVK <%s>: %s", gvk.String(), err.Error())
		} else {
			log.Info().Msgf("NetworkResourceWatchers: added event handler for GVK <%s>", gvk.String())
		}
	}
}

func isRunaiManagedNamespace(ctx context.Context, cachedClient client.Client, namespaceName string) bool {
	if namespaceName == "" {
		return false
	}
	namespace, err := utils.GetNamespace(ctx, namespaceName, cachedClient)
	if err != nil {
		return false
	}
	return utils.IsRunaiManagedNamespace(namespace)
}

func toClientObject(obj any) (client.Object, bool) {
	if o, ok := obj.(client.Object); ok {
		return o, true
	}
	tombstone, ok := obj.(clientgocache.DeletedFinalStateUnknown)
	if !ok {
		return nil, false
	}
	o, ok := tombstone.Obj.(client.Object)
	return o, ok
}
