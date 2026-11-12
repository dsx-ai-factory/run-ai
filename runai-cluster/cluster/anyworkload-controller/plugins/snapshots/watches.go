/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Watchers reconciles the runner when one of its replicas turns Ready, and when
// a capture changes. Without the pod watch the capture would depend on the
// workload's own status changing after the pod cache saw the replica Ready; the
// framework never requeues on a timer.
//
// Only with Snapshot installed: a watch on a kind whose CRD is missing stops the
// controller starting.
func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	if !p.snapshotInstalled {
		return bldr, nil
	}
	toRunner := handler.EnqueueRequestsFromMapFunc(mapToRunner)
	return bldr.
		Watches(&snapshotv1alpha1.SnapshotJob{}, toRunner).
		Watches(&corev1.Pod{}, toRunner, builder.WithPredicates(replicaReadinessChanged())), nil
}

// mapToRunner reads the runner's name from its label: MutateResource stamps it
// on replicas, and buildSnapshotJob sets it on the SnapshotJob and its PodSnapshot.
func mapToRunner(_ context.Context, object client.Object) []ctrlreconcile.Request {
	name := object.GetLabels()[commonconstants.WorkloadRunnerNameLabel]
	if name == "" {
		return nil
	}
	return []ctrlreconcile.Request{{NamespacedName: types.NamespacedName{Namespace: object.GetNamespace(), Name: name}}}
}

// replicaReadinessChanged passes opted-in replicas whose Ready condition changed,
// and on a restart those that are already Ready.
func replicaReadinessChanged() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			pod, ok := asReplica(e.Object)
			return ok && podReady(pod)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			before, wasReplica := asReplica(e.ObjectOld)
			after, isReplica := asReplica(e.ObjectNew)
			return wasReplica && isReplica && podReady(before) != podReady(after)
		},
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

func asReplica(object client.Object) (*corev1.Pod, bool) {
	pod, ok := object.(*corev1.Pod)
	if !ok || pod.Labels[constants.SnapshotOptInLabel] != "true" ||
		pod.Labels[commonconstants.WorkloadRunnerNameLabel] == "" {
		return nil, false
	}
	return pod, true
}
