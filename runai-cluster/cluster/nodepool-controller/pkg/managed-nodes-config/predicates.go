/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managed_nodes_config

import (
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type MNCPredicate struct {
}

func (M MNCPredicate) Create(e event.TypedCreateEvent[*v1alpha1.ManagedNodesConfig]) bool {
	return true
}

func (M MNCPredicate) Delete(e event.TypedDeleteEvent[*v1alpha1.ManagedNodesConfig]) bool {
	return true
}

func (M MNCPredicate) Update(e event.TypedUpdateEvent[*v1alpha1.ManagedNodesConfig]) bool {
	return filterMNCUpdates(e.ObjectOld, e.ObjectNew)
}

func (M MNCPredicate) Generic(e event.TypedGenericEvent[*v1alpha1.ManagedNodesConfig]) bool {
	return true
}

type NodePredicate struct {
}

func (M NodePredicate) Create(e event.TypedCreateEvent[*corev1.Node]) bool {
	return true
}

func (M NodePredicate) Delete(e event.TypedDeleteEvent[*corev1.Node]) bool {
	return true
}

func (M NodePredicate) Update(e event.TypedUpdateEvent[*corev1.Node]) bool {
	return filterNodeUpdatesForMNCController(e.ObjectOld, e.ObjectNew)
}

func (M NodePredicate) Generic(e event.TypedGenericEvent[*corev1.Node]) bool {
	return true
}

func filterNodeUpdatesForMNCController(nodeOld, nodeNew *corev1.Node) bool {
	return !reflect.DeepEqual(nodeOld.Labels, nodeNew.Labels) || nodeOld.Generation != nodeNew.Generation
}

func filterMNCUpdates(old, new *v1alpha1.ManagedNodesConfig) bool {
	return new.GetName() == config.Get().ManagedNodesConfigName && old.Generation != new.Generation
}
