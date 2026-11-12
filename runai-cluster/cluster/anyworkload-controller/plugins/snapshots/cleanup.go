/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"fmt"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NewCleanupSetup returns the manager setup that deletes a workload's
// SnapshotJobs and PodSnapshots once its runner is gone. The plugin can't: dromos
// runs no plugin for a deleted runner.
//
// Only with Snapshot installed, as for the plugin's watches.
func NewCleanupSetup() func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, mgr ctrl.Manager) error {
		toRunner := handler.EnqueueRequestsFromMapFunc(mapToRunner)
		return ctrl.NewControllerManagedBy(mgr).
			Named("snapshot-cleanup").
			Watches(&dromosv1alpha1.WorkloadRunner{}, &handler.EnqueueRequestForObject{},
				builder.WithPredicates(onlyDeletes())).
			// Catches a runner deleted while the controller was down.
			Watches(&snapshotv1alpha1.SnapshotJob{}, toRunner, builder.WithPredicates(onlyCreates())).
			Watches(&snapshotv1alpha1.PodSnapshot{}, toRunner, builder.WithPredicates(onlyCreates())).
			Complete(&cleaner{client: mgr.GetClient(), apiReader: mgr.GetAPIReader()})
	}
}

type cleaner struct {
	client    client.Client
	apiReader client.Reader
}

// Reconcile deletes what was made for a runner of this name other than the
// current one.
func (r *cleaner) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	made, err := madeFor(ctx, r.client, request.NamespacedName)
	if err != nil {
		return ctrl.Result{}, err
	}

	// After the listing, and uncached, so the runner is never older than what
	// was listed: a stale "not found" or UID would delete a live runner's captures.
	current := &dromosv1alpha1.WorkloadRunner{}
	if err := r.apiReader.Get(ctx, request.NamespacedName, current); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("get runner: %w", err)
	}
	for _, object := range made {
		// current.UID is empty when no runner has this name, so all of it goes.
		if object.GetLabels()[commonconstants.WorkloadRunnerUidLabel] == string(current.UID) {
			continue
		}
		if err := r.client.Delete(ctx, object); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("delete %T %s: %w", object, object.GetName(), err)
		}
	}
	return ctrl.Result{}, nil
}

// madeFor lists the SnapshotJobs and PodSnapshots made for runners of this name.
func madeFor(ctx context.Context, c client.Reader, name types.NamespacedName) ([]client.Object, error) {
	jobs := &snapshotv1alpha1.SnapshotJobList{}
	snapshots := &snapshotv1alpha1.PodSnapshotList{}
	for _, list := range []client.ObjectList{jobs, snapshots} {
		// Requiring the UID leaves alone what this plugin didn't make.
		if err := c.List(ctx, list, client.InNamespace(name.Namespace),
			client.MatchingLabels{commonconstants.WorkloadRunnerNameLabel: name.Name},
			client.HasLabels{commonconstants.WorkloadRunnerUidLabel}); err != nil {
			return nil, fmt.Errorf("list %T: %w", list, err)
		}
	}

	var made []client.Object
	for i := range jobs.Items {
		made = append(made, &jobs.Items[i])
	}
	for i := range snapshots.Items {
		made = append(made, &snapshots.Items[i])
	}
	return made, nil
}

func onlyDeletes() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return false },
		UpdateFunc:  func(event.UpdateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

func onlyCreates() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		UpdateFunc:  func(event.UpdateEvent) bool { return false },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}
