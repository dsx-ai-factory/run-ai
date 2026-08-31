/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package affinity holds the node-affinity merge primitives shared by the dromos
// plugins that write scheduling constraints onto a workload's pods. Every merge
// here is idempotent, which is what keeps the framework's multi-round
// MutateResource loop converging.
package affinity

import (
	"slices"

	corev1 "k8s.io/api/core/v1"
)

// MergeNodeAffinity merges src into dst in place.
func MergeNodeAffinity(dst, src *corev1.NodeAffinity) {
	if dst == nil || src == nil {
		return
	}

	dst.RequiredDuringSchedulingIgnoredDuringExecution = MergeRequired(
		dst.RequiredDuringSchedulingIgnoredDuringExecution,
		src.RequiredDuringSchedulingIgnoredDuringExecution,
	)
	dst.PreferredDuringSchedulingIgnoredDuringExecution = MergePreferred(
		dst.PreferredDuringSchedulingIgnoredDuringExecution,
		src.PreferredDuringSchedulingIgnoredDuringExecution,
	)
}

// MergeRequired ANDs two required-node-affinity NodeSelectors. Terms are OR-ed by
// the scheduler, so the only way to AND two selectors is the cartesian product:
//
//	first:  [termA, termB]
//	second: [termC, termD]
//	merged: [termA && termC, termA && termD, termB && termC, termB && termD]
//
// The Implies short-circuit is load-bearing, not an optimization: without it a
// re-merge would square the term count every round and the mutation loop would
// hit its round cap, leaving the child manifest unapplied.
func MergeRequired(first, second *corev1.NodeSelector) *corev1.NodeSelector {
	// A matches-nothing selector wins the merge: ANDing with it still matches nothing.
	if matchesNothing(first) || matchesNothing(second) {
		return &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{}}}
	}

	// An empty term matches no nodes, so it contributes nothing to the OR its
	// selector expresses. Upstream drops such terms; do the same, or the
	// containment test below would read one as "satisfied by anything" and skip
	// the merge entirely, silently discarding the rest of the selector.
	first, second = withoutEmptyTerms(first), withoutEmptyTerms(second)

	switch {
	case isUnconstrained(first) && isUnconstrained(second):
		return nil
	case isUnconstrained(first):
		return second.DeepCopy()
	case isUnconstrained(second):
		return first.DeepCopy()
	case Implies(first, second):
		return first.DeepCopy()
	}

	merged := &corev1.NodeSelector{
		NodeSelectorTerms: make([]corev1.NodeSelectorTerm, 0, len(first.NodeSelectorTerms)*len(second.NodeSelectorTerms)),
	}
	for _, firstTerm := range first.NodeSelectorTerms {
		for _, secondTerm := range second.NodeSelectorTerms {
			merged.NodeSelectorTerms = append(merged.NodeSelectorTerms, MergeTerms(firstTerm, secondTerm))
		}
	}

	return merged
}

// Implies reports whether every node matching first also matches second, so that
// ANDing the two would yield first unchanged.
//
// Matching is by requirement equality rather than by evaluating the selectors, so
// the answer is only ever wrong in the safe direction: it may return false for a
// pair that does imply, which costs one redundant merge and nothing else.
func Implies(first, second *corev1.NodeSelector) bool {
	if second == nil || len(second.NodeSelectorTerms) == 0 {
		return true
	}
	if first == nil || len(first.NodeSelectorTerms) == 0 {
		return false
	}

	for _, firstTerm := range first.NodeSelectorTerms {
		satisfiesSomeTerm := slices.ContainsFunc(second.NodeSelectorTerms, func(secondTerm corev1.NodeSelectorTerm) bool {
			return termContains(firstTerm, secondTerm)
		})
		if !satisfiesSomeTerm {
			return false
		}
	}
	return true
}

// termContains reports whether superset carries every requirement of subset. An
// empty subset is never contained: it matches no nodes, so nothing satisfies it.
func termContains(superset, subset corev1.NodeSelectorTerm) bool {
	if isEmptyTerm(subset) {
		return false
	}
	return requirementsContain(superset.MatchExpressions, subset.MatchExpressions) &&
		requirementsContain(superset.MatchFields, subset.MatchFields)
}

func isEmptyTerm(term corev1.NodeSelectorTerm) bool {
	return len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0
}

func isUnconstrained(selector *corev1.NodeSelector) bool {
	return selector == nil || len(selector.NodeSelectorTerms) == 0
}

// matchesNothing reports whether selector has terms and every one is empty -
// such a selector matches no nodes.
func matchesNothing(selector *corev1.NodeSelector) bool {
	if isUnconstrained(selector) {
		return false
	}
	return !slices.ContainsFunc(selector.NodeSelectorTerms, func(term corev1.NodeSelectorTerm) bool {
		return !isEmptyTerm(term)
	})
}

// withoutEmptyTerms returns selector with its no-op terms removed, or selector
// itself when there are none.
func withoutEmptyTerms(selector *corev1.NodeSelector) *corev1.NodeSelector {
	if selector == nil || !slices.ContainsFunc(selector.NodeSelectorTerms, isEmptyTerm) {
		return selector
	}
	kept := make([]corev1.NodeSelectorTerm, 0, len(selector.NodeSelectorTerms))
	for _, term := range selector.NodeSelectorTerms {
		if !isEmptyTerm(term) {
			kept = append(kept, term)
		}
	}
	return &corev1.NodeSelector{NodeSelectorTerms: kept}
}

func requirementsContain(superset, subset []corev1.NodeSelectorRequirement) bool {
	for _, req := range subset {
		present := slices.ContainsFunc(superset, func(have corev1.NodeSelectorRequirement) bool {
			return have.Key == req.Key &&
				have.Operator == req.Operator &&
				stringSlicesEquivalent(have.Values, req.Values)
		})
		if !present {
			return false
		}
	}
	return true
}

// MergePreferred appends src's terms to dst, skipping duplicates. Preferred terms
// are independent, so there is no cartesian product. Duplicates compare weight
// and requirement values as sets.
func MergePreferred(dst, src []corev1.PreferredSchedulingTerm) []corev1.PreferredSchedulingTerm {
	for _, term := range src {
		alreadyPresent := slices.ContainsFunc(dst, func(seen corev1.PreferredSchedulingTerm) bool {
			return seen.Weight == term.Weight && termsEquivalent(seen.Preference, term.Preference)
		})
		if !alreadyPresent {
			dst = append(dst, *term.DeepCopy())
		}
	}
	return dst
}

func termsEquivalent(a, b corev1.NodeSelectorTerm) bool {
	return requirementsEquivalent(a.MatchExpressions, b.MatchExpressions) &&
		requirementsEquivalent(a.MatchFields, b.MatchFields)
}

func requirementsEquivalent(a, b []corev1.NodeSelectorRequirement) bool {
	return requirementsContain(a, b) && requirementsContain(b, a)
}

// NormalizeNodeAffinity sorts and de-duplicates every requirement's values in
// place, so a caller assembling them from an unordered source still yields a
// byte-stable child manifest.
func NormalizeNodeAffinity(na *corev1.NodeAffinity) {
	if na == nil {
		return
	}
	if required := na.RequiredDuringSchedulingIgnoredDuringExecution; required != nil {
		for i := range required.NodeSelectorTerms {
			normalizeTerm(&required.NodeSelectorTerms[i])
		}
	}
	for i := range na.PreferredDuringSchedulingIgnoredDuringExecution {
		normalizeTerm(&na.PreferredDuringSchedulingIgnoredDuringExecution[i].Preference)
	}
}

func normalizeTerm(term *corev1.NodeSelectorTerm) {
	for _, requirements := range [][]corev1.NodeSelectorRequirement{term.MatchExpressions, term.MatchFields} {
		for i := range requirements {
			slices.Sort(requirements[i].Values)
			requirements[i].Values = slices.Compact(requirements[i].Values)
		}
	}
}

// MergeTerms AND-merges two NodeSelectorTerms, dropping requirements from second
// that first already expresses.
func MergeTerms(first, second corev1.NodeSelectorTerm) corev1.NodeSelectorTerm {
	return corev1.NodeSelectorTerm{
		MatchExpressions: mergeRequirements(first.MatchExpressions, second.MatchExpressions),
		MatchFields:      mergeRequirements(first.MatchFields, second.MatchFields),
	}
}

func mergeRequirements(first, second []corev1.NodeSelectorRequirement) []corev1.NodeSelectorRequirement {
	var merged []corev1.NodeSelectorRequirement
	byKey := make(map[string][]corev1.NodeSelectorRequirement)

	for _, requirements := range [][]corev1.NodeSelectorRequirement{first, second} {
		for _, req := range requirements {
			if slices.ContainsFunc(byKey[req.Key], func(seen corev1.NodeSelectorRequirement) bool {
				return req.Operator == seen.Operator && stringSlicesEquivalent(req.Values, seen.Values)
			}) {
				continue
			}
			// Copied: a cartesian merge emits one requirement into many terms, and
			// a shared Values array would let a later sort on one reorder the rest.
			merged = append(merged, *req.DeepCopy())
			byKey[req.Key] = append(byKey[req.Key], req)
		}
	}

	return merged
}

func stringSlicesEquivalent(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int)
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}
