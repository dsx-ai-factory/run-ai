/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"fmt"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HandleRunnerDeleted deletes a workload's SnapshotJobs and PodSnapshots once its
// runner is gone. Garbage collection can't: the runner doesn't own them, since a
// SnapshotJob it owned would show as a second workload of the same runner.
func (p *Plugin) HandleRunnerDeleted(ctx context.Context, runner types.NamespacedName, k8sClient client.Client) error {
	// Listing a kind whose CRD is missing fails.
	if !p.snapshotInstalled {
		return nil
	}
	return deleteOthers(ctx, k8sClient, runner, "")
}

// deleteOthers deletes what was made for a runner of this name other than
// current, the UID of the runner being reconciled, or empty when it is gone.
//
// current is the caller's, from the cache. Nothing else captures for this name
// during its reconcile, so only a cache still stale after a restart could take a
// live runner's captures for another's, and those are captured again.
func deleteOthers(ctx context.Context, c client.Client, name types.NamespacedName, current types.UID) error {
	runnersSnapshots, err := getRunnersSnapshots(ctx, c, name)
	if err != nil {
		return err
	}
	for _, object := range runnersSnapshots {
		if object.GetLabels()[commonconstants.WorkloadRunnerUidLabel] == string(current) {
			continue
		}
		if err := c.Delete(ctx, object); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete %T %s: %w", object, object.GetName(), err)
		}
	}
	return nil
}

// getRunnersSnapshots lists the SnapshotJobs and PodSnapshots made for runners of this name.
func getRunnersSnapshots(ctx context.Context, c client.Reader, name types.NamespacedName) ([]client.Object, error) {
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

	var runnersSnapshots []client.Object
	for i := range jobs.Items {
		runnersSnapshots = append(runnersSnapshots, &jobs.Items[i])
	}
	for i := range snapshots.Items {
		runnersSnapshots = append(runnersSnapshots, &snapshots.Items[i])
	}
	return runnersSnapshots, nil
}
