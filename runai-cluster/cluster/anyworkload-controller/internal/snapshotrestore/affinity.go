/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"fmt"
	"maps"
	"slices"

	corev1 "k8s.io/api/core/v1"
)

// requireNodeSelector ANDs selector into the pod's required node affinity, so
// the replica can only land where Snapshot's agent runs.
//
// Expressions are added to every existing term rather than as a new term: terms
// are ORed, so a new one would widen placement instead of narrowing it. The
// workload's own expressions are kept and ANDed with ours - two In requirements
// on a key intersect, which is what both constraints together mean.
//
// Nothing is removed. If any term contradicts the agent, the pod cannot both
// honour its placement and restore, so an error is returned and the caller
// cold-starts it rather than admitting one that would stay Pending.
//
// Not admission-controller's patch.AddNodeAffinityToPodSpec: that appends one
// term per expression, which ORs a multi-key selector against itself.
func requireNodeSelector(spec *corev1.PodSpec, selector map[string]string) error {
	// Sorted so an unchanged selector produces an unchanged pod.
	required := make([]corev1.NodeSelectorRequirement, 0, len(selector))
	for _, key := range slices.Sorted(maps.Keys(selector)) {
		// spec.nodeSelector is ANDed with affinity by the scheduler and cannot
		// be narrowed from here, so a conflict there is a conflict outright.
		if value, set := spec.NodeSelector[key]; set && value != selector[key] {
			return fmt.Errorf("%w: nodeSelector pins %s=%s, the agent runs on %s=%s",
				errUnplaceable, key, value, key, selector[key])
		}
		required = append(required, corev1.NodeSelectorRequirement{
			Key: key, Operator: corev1.NodeSelectorOpIn, Values: []string{selector[key]},
		})
	}
	if len(required) == 0 {
		return nil
	}

	nodeSelector := requiredNodeAffinity(spec)
	if len(nodeSelector.NodeSelectorTerms) == 0 {
		// An empty term matches every node, so the loop below fills it in.
		nodeSelector.NodeSelectorTerms = []corev1.NodeSelectorTerm{{}}
	}

	// Nothing the workload asked for is removed: a term that contradicts the
	// agent means this pod cannot both honour its placement and restore, so it
	// cold-starts instead.
	for i, term := range nodeSelector.NodeSelectorTerms {
		if key, conflict := conflictsWith(term, selector); conflict {
			return fmt.Errorf("%w: node affinity on %s excludes %s=%s",
				errUnplaceable, key, key, selector[key])
		}
		nodeSelector.NodeSelectorTerms[i].MatchExpressions =
			append(nodeSelector.NodeSelectorTerms[i].MatchExpressions, required...)
	}
	return nil
}

// requiredNodeAffinity returns the pod's required node selector, creating the
// chain of optional structs down to it.
func requiredNodeAffinity(spec *corev1.PodSpec) *corev1.NodeSelector {
	if spec.Affinity == nil {
		spec.Affinity = &corev1.Affinity{}
	}
	if spec.Affinity.NodeAffinity == nil {
		spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	affinity := spec.Affinity.NodeAffinity
	if affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}
	return affinity.RequiredDuringSchedulingIgnoredDuringExecution
}

// conflictsWith reports the first key on which term can never match a node the
// agent runs on. Only operators that can contradict an In of one value are
// considered; the rest leave the term satisfiable.
func conflictsWith(term corev1.NodeSelectorTerm, selector map[string]string) (string, bool) {
	for _, expression := range term.MatchExpressions {
		value, constrained := selector[expression.Key]
		if !constrained {
			continue
		}

		switch expression.Operator {
		case corev1.NodeSelectorOpIn:
			if !contains(expression.Values, value) {
				return expression.Key, true
			}
		case corev1.NodeSelectorOpNotIn:
			if contains(expression.Values, value) {
				return expression.Key, true
			}
		case corev1.NodeSelectorOpDoesNotExist:
			return expression.Key, true
		}
	}
	return "", false
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
