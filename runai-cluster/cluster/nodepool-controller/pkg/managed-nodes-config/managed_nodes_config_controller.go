/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managed_nodes_config

import (
	"context"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ManagedNodesConfigController struct {
	Client client.Client
	Scheme *runtime.Scheme

	*nodepool_controller.NodePoolController
}

type MNCReconcileRequest struct {
	types.NamespacedName
	isNode bool
}

func NewManagedNodesConfigController(client client.Client, scheme *runtime.Scheme, nodePoolController *nodepool_controller.NodePoolController) *ManagedNodesConfigController {
	return &ManagedNodesConfigController{
		Client:             client,
		Scheme:             scheme,
		NodePoolController: nodePoolController,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (mncc *ManagedNodesConfigController) SetupWithManager(_ context.Context, mgr ctrl.Manager) error {
	return builder.TypedControllerManagedBy[MNCReconcileRequest](mgr).
		Named("managed-nodes-config-controller").
		WatchesRawSource(source.TypedKind(
			mgr.GetCache(),
			&v1alpha1.ManagedNodesConfig{},
			handler.TypedEnqueueRequestsFromMapFunc(mncc.MapManagedNodesConfigToEvent),
			MNCPredicate{},
		)).
		WatchesRawSource(source.TypedKind(
			mgr.GetCache(),
			&corev1.Node{},
			handler.TypedEnqueueRequestsFromMapFunc(mncc.MapNodeToManagedNodesConfigEvent),
			NodePredicate{},
		)).
		WithOptions(controller.TypedOptions[MNCReconcileRequest]{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[MNCReconcileRequest](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(mncc)
}

func (mncc *ManagedNodesConfigController) MapNodeToManagedNodesConfigEvent(_ context.Context, node *corev1.Node) (requests []MNCReconcileRequest) {
	requests = append(requests, MNCReconcileRequest{
		NamespacedName: types.NamespacedName{
			Name: node.Name},
		isNode: true,
	})
	return requests
}

func (mncc *ManagedNodesConfigController) MapManagedNodesConfigToEvent(_ context.Context, mnc *v1alpha1.ManagedNodesConfig) (requests []MNCReconcileRequest) {
	requests = append(requests, MNCReconcileRequest{
		NamespacedName: types.NamespacedName{
			Name: mnc.Name},
		isNode: false,
	})
	return requests
}
