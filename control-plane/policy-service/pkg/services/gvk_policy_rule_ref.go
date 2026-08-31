/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// PolicyRuleRefFromRule returns the stable identity of a full policy rule (without enforcements).
func PolicyRuleRefFromRule(r api.PolicyRule) api.PolicyRuleRef {
	return api.PolicyRuleRef{
		Field:          r.Field,
		SpecSelector:   r.SpecSelector,
		ItemSelector:   r.ItemSelector,
		SelectorParams: r.SelectorParams,
	}
}

const deletedSentinel = -1

// PolicyRuleSet stores policy rules in an ordered slice with a map index for O(1)
// lookup by ref key. Deleted slots are marked with a sentinel index and skipped
// when producing the final Rules() slice.
type PolicyRuleSet struct {
	rules []api.PolicyRule
	index map[string]int // ref key -> position in rules (deletedSentinel = removed)
}

func NewPolicyRuleSet(rules []api.PolicyRule) *PolicyRuleSet {
	s := &PolicyRuleSet{
		rules: make([]api.PolicyRule, len(rules)),
		index: make(map[string]int, len(rules)),
	}
	copy(s.rules, rules)
	for i, r := range s.rules {
		key := dbmodel.PolicyRuleRefKey(PolicyRuleRefFromRule(r))
		s.index[key] = i
	}
	return s
}

// Set adds or replaces a rule. If a rule with the same ref already exists it is
// replaced in-place (preserving its position). Otherwise the rule is appended.
func (s *PolicyRuleSet) Set(rule api.PolicyRule) {
	key := dbmodel.PolicyRuleRefKey(PolicyRuleRefFromRule(rule))
	if idx, ok := s.index[key]; ok && idx != deletedSentinel {
		s.rules[idx] = rule
		return
	}
	s.index[key] = len(s.rules)
	s.rules = append(s.rules, rule)
}

// Delete removes a rule by ref. Returns false if the ref was not found.
func (s *PolicyRuleSet) Delete(ref api.PolicyRuleRef) bool {
	key := dbmodel.PolicyRuleRefKey(ref)
	idx, ok := s.index[key]
	if !ok || idx == deletedSentinel {
		return false
	}
	s.index[key] = deletedSentinel
	return true
}

// Has returns true if the set contains a rule matching the given ref.
func (s *PolicyRuleSet) Has(ref api.PolicyRuleRef) bool {
	strKey := dbmodel.PolicyRuleRefKey(ref)
	idx, ok := s.index[strKey]
	return ok && idx != deletedSentinel
}

// Rules returns the ordered slice of rules, skipping any deleted entries.
func (s *PolicyRuleSet) Rules() []api.PolicyRule {
	out := make([]api.PolicyRule, 0, len(s.rules))

	for i, r := range s.rules {
		key := dbmodel.PolicyRuleRefKey(PolicyRuleRefFromRule(r))
		if idx, ok := s.index[key]; ok && idx == i {
			out = append(out, r)
		}
	}

	return out
}

// mergeGvkPolicyRulesByRef updates rules whose ref matches an incoming rule (full replacement of that rule),
// and appends rules that do not match any existing ref. Order is stable: unmatched existing rules keep order,
// then new rules are appended in request order.
func mergeGvkPolicyRulesByRef(existing []api.PolicyRule, incoming []api.PolicyRule) []api.PolicyRule {
	if len(incoming) == 0 {
		return existing
	}

	set := NewPolicyRuleSet(existing)

	for _, in := range incoming {
		set.Set(in)
	}

	return set.Rules()
}

// removeGvkPolicyRulesByRefs removes every rule whose ref matches any entry in toRemove.
func removeGvkPolicyRulesByRefs(rules []api.PolicyRule, toRemove []api.PolicyRuleRef) ([]api.PolicyRule, *api.Error) {
	if len(toRemove) == 0 {
		return rules, nil
	}

	set := NewPolicyRuleSet(rules)

	for _, rm := range toRemove {
		if !set.Has(rm) {
			return nil, api.NewValidationError(fmt.Sprintf("rule not found: %s", dbmodel.PolicyRuleRefKey(rm)))
		}

		set.Delete(rm)
	}

	return set.Rules(), nil
}
