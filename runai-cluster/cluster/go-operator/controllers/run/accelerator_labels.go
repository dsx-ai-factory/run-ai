/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"maps"
	"strings"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// KAI accounts for AMD and NVIDIA whole GPUs in a single GPU slot. Publish a
// placement label from each advertised extended resource so a workload asking
// for a particular resource can be matched to the right node as well as quota.
func SetupAcceleratorNodeLabels(manager ctrl.Manager) error {
	c := manager.GetClient()
	needsUpdate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		_, changed := acceleratorNodeLabels(obj.(*corev1.Node))
		return changed
	})
	return ctrl.NewControllerManagedBy(manager).
		Named("accelerator-node-labels").
		For(&corev1.Node{}).
		WithEventFilter(needsUpdate).
		Complete(reconcile.Func(func(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
			var node corev1.Node
			if err := c.Get(ctx, request.NamespacedName, &node); err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			labels, changed := acceleratorNodeLabels(&node)
			if !changed {
				return ctrl.Result{}, nil
			}
			before := client.MergeFromWithOptions(node.DeepCopy(), client.MergeFromWithOptimisticLock{})
			node.Labels = labels
			return ctrl.Result{}, c.Patch(ctx, &node, before)
		}))
}

func acceleratorNodeLabels(node *corev1.Node) (map[string]string, bool) {
	desired := make(map[string]string)
	if node.Labels[constants.NodePoolNameLabel] != clusterconstants.ExcludedNodepoolName {
		for name, quantity := range node.Status.Allocatable {
			// Native Kubernetes resources, including attach limits, are not accelerators.
			domain, _, qualified := strings.Cut(string(name), "/")
			if !qualified || domain == "kubernetes.io" || strings.HasSuffix(domain, ".kubernetes.io") || quantity.Sign() <= 0 {
				continue
			}
			desired[constants.AcceleratorResourceNodeLabel(string(name))] = "true"
		}
	}

	current := make(map[string]string)
	for key, value := range node.Labels {
		if strings.HasPrefix(key, constants.AcceleratorResourceLabelPrefix) {
			current[key] = value
		}
	}
	if maps.Equal(desired, current) {
		return nil, false
	}
	labels := maps.Clone(node.Labels)
	if labels == nil {
		labels = make(map[string]string, len(desired))
	}
	for key := range current {
		delete(labels, key)
	}
	maps.Copy(labels, desired)
	return labels, true
}
