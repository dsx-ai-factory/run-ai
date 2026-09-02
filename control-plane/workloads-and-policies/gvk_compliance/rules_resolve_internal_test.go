/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

// Serial: BrokenRuleHook is package-global; these specs must not run in parallel
// with each other or with anything else that mutates the hook.
var _ = Describe("ResolveRules and findFieldVariation", Serial, func() {
	var msgs []string

	BeforeEach(func() {
		msgs = nil
		BrokenRuleHook = func(_ *policyApi.EffectiveRulesEffectiveRulesInner, msg string) {
			msgs = append(msgs, msg)
		}
	})

	AfterEach(func() {
		BrokenRuleHook = nil
	})

	It("logs a broken rule when the spec selector has no query", func() {
		pol := minimalPolicyForResolveWithSpecSelector(policyApi.BreakdownSpecSelector{Query: nil})
		_ = ResolveRules(&pol)
		Expect(msgs).To(HaveLen(1))
		Expect(msgs[0]).To(ContainSubstring("has no query"))
	})

	It("logs a broken rule when the composed selector JQ fails to compile", func() {
		pol := minimalPolicyForResolve("self", "..this is not jq")
		_ = ResolveRules(&pol)
		Expect(msgs).To(HaveLen(1))
		Expect(msgs[0]).To(ContainSubstring("failed to compile selector query"))
	})

	It("findFieldVariation logs when breakdown.Fields is nil", func() {
		rule := policyApi.EffectiveRulesEffectiveRulesInner{
			Field:        "cpu",
			SpecSelector: "self",
			Enforcements: []policyApi.PolicyRuleEnforcement{minimalRangeEnforcement()},
		}
		bd := policyApi.BreakdownBreakdown{Fields: nil}
		Expect(findFieldVariation(&rule, &bd)).To(BeNil())
		Expect(msgs).To(Equal([]string{"breakdown has no fields"}))
	})
})

func minimalRangeEnforcement() policyApi.PolicyRuleEnforcement {
	rng := policyApi.NewPolicyEnforcementRangeRange()
	rng.SetMax(5)
	par := policyApi.NewPolicyEnforcementRange()
	par.SetRange(*rng)
	return policyApi.PolicyRuleEnforcement{PolicyEnforcementRange: par}
}

func minimalPolicyForResolve(specSelectorName, selectorJQ string) policyApi.GetGvkPolicyResp {
	sel := policyApi.NewBreakdownSpecSelector()
	sel.SetQuery(selectorJQ)
	return minimalPolicyForResolveWithSpecSelectorKey(specSelectorName, *sel)
}

func minimalPolicyForResolveWithSpecSelector(sel policyApi.BreakdownSpecSelector) policyApi.GetGvkPolicyResp {
	return minimalPolicyForResolveWithSpecSelectorKey("self", sel)
}

func minimalPolicyForResolveWithSpecSelectorKey(specSelectorName string, sel policyApi.BreakdownSpecSelector) policyApi.GetGvkPolicyResp {
	f := policyApi.NewBreakdownFormatter()
	f.SetBuiltin(KeyValueBuiltin)
	fields := map[string][]policyApi.BreakdownFieldVariation{
		"cpu": {{
			SpecSelectors: []string{specSelectorName},
			SpecRef:       convutils.Ref("cpu"),
			Formatter:     f,
		}},
	}
	specSelectors := map[string]policyApi.BreakdownSpecSelector{
		specSelectorName: sel,
	}
	bd := policyApi.BreakdownBreakdown{}
	bd.SetFields(fields)
	bd.SetSpecSelectors(specSelectors)

	pol := policyApi.GetGvkPolicyResp{}
	pol.SetId("unit-test")
	pol.SetBreakdown(bd)
	pol.SetEffectiveRules([]policyApi.EffectiveRulesEffectiveRulesInner{{
		Field:        "cpu",
		SpecSelector: specSelectorName,
		Enforcements: []policyApi.PolicyRuleEnforcement{minimalRangeEnforcement()},
	}})
	return pol
}
