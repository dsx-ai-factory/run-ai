/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeaffinity

import (
	"context"
	"errors"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	affinityutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/affinity"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var errTooManyMergedTerms = errors.New("merging the node affinity instruction exceeds the term limit")

// MutateResource merges the runner's node-affinity instruction into the pod
// affinity of every component of the rendered child.
func (p *Plugin) MutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	desired := extractNodeAffinity(runner)
	if desired == nil {
		return nil
	}

	if err := affinityutils.SupportsNodeAffinity(karta); err != nil {
		return state.AddPluginIssue(ctx, p, resource,
			fmt.Errorf("karta %s does not support node affinity: %w", karta.Name, err), true, "", issue.Compute)
	}

	desired = desired.DeepCopy()
	affinityutils.NormalizeNodeAffinity(desired)

	mutate := func(instance kartaresource.ExtractedInstance) error {
		return mergeIntoInstance(instance, desired)
	}
	if err := kartautils.MutatePodInstances(ctx, karta, resource, mutate); err != nil {
		group := issue.Internal
		if errors.Is(err, errTooManyMergedTerms) {
			group = issue.Compute
		}
		return state.AddPluginIssue(ctx, p, resource, err, true, "", group)
	}

	return nil
}

// mergeIntoInstance merges desired into the instance's node affinity, whichever
// shape Karta extracted the pod spec into.
func mergeIntoInstance(instance kartaresource.ExtractedInstance, desired *corev1.NodeAffinity) error {
	var target *corev1.NodeAffinity

	switch {
	case instance.PodTemplateSpec != nil:
		target = ensurePodSpecNodeAffinity(&instance.PodTemplateSpec.Spec)
	case instance.PodSpec != nil:
		target = ensurePodSpecNodeAffinity(instance.PodSpec)
	case instance.FragmentedPodSpec != nil:
		target = ensureFragmentedNodeAffinity(instance.FragmentedPodSpec)
	default:
		return nil
	}

	if merged := requiredTermCount(target) * requiredTermCount(desired); merged > maxMergedTerms {
		return fmt.Errorf("%w: %d terms, max %d", errTooManyMergedTerms, merged, maxMergedTerms)
	}

	// Copied per instance so no two pod specs share a selector.
	affinityutils.MergeNodeAffinity(target, desired.DeepCopy())
	return nil
}

func requiredTermCount(affinity *corev1.NodeAffinity) int {
	if affinity == nil || affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return 0
	}
	return len(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)
}

func ensurePodSpecNodeAffinity(podSpec *corev1.PodSpec) *corev1.NodeAffinity {
	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{}
	}
	if podSpec.Affinity.NodeAffinity == nil {
		podSpec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	return podSpec.Affinity.NodeAffinity
}

func ensureFragmentedNodeAffinity(frag *kartaresource.FragmentedPodSpec) *corev1.NodeAffinity {
	if frag.NodeAffinity == nil {
		frag.NodeAffinity = &corev1.NodeAffinity{}
	}
	return frag.NodeAffinity
}
