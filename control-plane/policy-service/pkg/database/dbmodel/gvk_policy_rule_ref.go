/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"fmt"
	"slices"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// JSON field names (matching the JSON tags on api.PolicyRule /
// api.EffectiveRulesEffectiveRulesInner) used to identify a GVK policy rule
// when it is held as a generic map[string]interface{}.
const (
	gvkRuleFieldKey          = "field"
	gvkRuleSpecSelectorKey   = "specSelector"
	gvkRuleItemSelectorKey   = "itemSelector"
	gvkRuleSelectorParamsKey = "selectorParams"
)

// PolicyRuleRefKey produces a canonical, deterministic string key for a
// PolicyRuleRef. Two refs that target the same field with the same spec
// selector, item selector and selector params produce the same key, regardless
// of the iteration order of the selector params map. The key is used as the
// stable identity of a rule for lookup and merging across policies/scopes.
func PolicyRuleRefKey(ref api.PolicyRuleRef) string {
	key := fmt.Sprintf("%s->%s", ref.Field, ref.SpecSelector)

	if ref.ItemSelector != nil {
		key += "->" + *ref.ItemSelector
	}

	if len(ref.SelectorParams) == 0 {
		return key
	}

	params := make([]string, 0, len(ref.SelectorParams))
	for k, v := range ref.SelectorParams {
		params = append(params, fmt.Sprintf("[%s:%v]", k, v))
	}

	slices.Sort(params)

	return key + strings.Join(params, ";")
}

// PolicyRuleRefFromMap extracts the identity of a rule from its JSON map
// representation. Use this when rules are held as map[string]interface{} (for
// example inside dao_utils.JsonBMap) rather than as the typed api.PolicyRule
// struct.
func PolicyRuleRefFromMap(ruleMap map[string]interface{}) api.PolicyRuleRef {
	ref := api.PolicyRuleRef{}

	if v, ok := ruleMap[gvkRuleFieldKey].(string); ok {
		ref.Field = v
	}
	if v, ok := ruleMap[gvkRuleSpecSelectorKey].(string); ok {
		ref.SpecSelector = v
	}
	if v, ok := ruleMap[gvkRuleItemSelectorKey].(string); ok {
		ref.ItemSelector = &v
	}
	if v := convutils.MapMe(ruleMap[gvkRuleSelectorParamsKey]); v != nil {
		ref.SelectorParams = v
	}

	return ref
}
