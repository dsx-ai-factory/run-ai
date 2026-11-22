/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"fmt"
	"maps"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MutateResource stamps an eligible, opted-in workload's pod template with the
// opt-in, which the restore webhook selects pods by, and the runner's name and
// UID, by which the capture finds the workload's replicas. The UID tells them
// from those of an earlier runner of the same name.
//
// It does not check that Snapshot is installed, so the labels depend only on
// the workload and installing or removing Snapshot never rolls it.
func (p *Plugin) MutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	if !optedIn(resource) {
		return nil
	}
	if err := checkEligibility(ctx, resource, karta); err != nil {
		return nil //nolint:nilerr // Status reports it; an ineligible workload is applied unchanged.
	}

	labels := map[string]string{
		constants.SnapshotOptInLabel:            "true",
		commonconstants.WorkloadRunnerNameLabel: runner.Name,
		commonconstants.WorkloadRunnerUidLabel:  string(runner.UID),
	}
	err := kartautils.MutatePodInstances(ctx, karta, resource, func(instance kartaresource.ExtractedInstance) error {
		stampLabels(instance, labels)
		return nil
	})
	if err != nil {
		// A failed stamp means cold starts, never a workload that isn't applied.
		_ = state.AddPluginIssue(ctx, p, resource, fmt.Errorf("stamp pod labels: %w", err), false, "", issue.Internal)
	}
	return nil
}

// stampLabels adds labels to the instance's pod template, in whichever of the
// two shapes that carry labels Karta extracted it into.
func stampLabels(instance kartaresource.ExtractedInstance, labels map[string]string) {
	switch {
	case instance.PodTemplateSpec != nil:
		if instance.PodTemplateSpec.Labels == nil {
			instance.PodTemplateSpec.Labels = make(map[string]string, len(labels))
		}
		maps.Copy(instance.PodTemplateSpec.Labels, labels)
	case instance.FragmentedPodSpec != nil:
		if instance.FragmentedPodSpec.Labels == nil {
			instance.FragmentedPodSpec.Labels = make(map[string]string, len(labels))
		}
		maps.Copy(instance.FragmentedPodSpec.Labels, labels)
	}
}
