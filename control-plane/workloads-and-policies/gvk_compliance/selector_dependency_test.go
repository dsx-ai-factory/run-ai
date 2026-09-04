/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("selector_dependency", func() {
	Describe("ensurePath", func() {
		It("creates single-level path", func() {
			m := map[string]any{}
			ensurePath(m, []string{"foo"})
			Expect(m).To(HaveKey("foo"))
			Expect(m["foo"]).To(BeAssignableToTypeOf(map[string]any{}))
		})

		It("creates nested path", func() {
			m := map[string]any{}
			ensurePath(m, []string{"a", "b", "c"})
			Expect(m).To(HaveKey("a"))
			a := m["a"].(map[string]any)
			Expect(a).To(HaveKey("b"))
			b := a["b"].(map[string]any)
			Expect(b).To(HaveKey("c"))
		})

		It("preserves existing values", func() {
			m := map[string]any{
				"a": map[string]any{
					"existing": "value",
				},
			}
			ensurePath(m, []string{"a", "b"})
			a := m["a"].(map[string]any)
			Expect(a).To(HaveKey("existing"))
			Expect(a["existing"]).To(Equal("value"))
			Expect(a).To(HaveKey("b"))
		})

		It("does not overwrite non-map values", func() {
			m := map[string]any{
				"a": "string-value",
			}
			ensurePath(m, []string{"a", "b"})
			Expect(m["a"]).To(Equal("string-value"))
		})
	})

	Describe("SortDependencies", func() {
		It("returns nil for empty input", func() {
			sorted, err := SortDependencies(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(sorted).To(BeNil())
		})

		It("sorts dependencies with parent before child", func() {
			deps := []SpecSelectorDependency{
				{Name: "child", ParentName: "parent", ElementPath: []string{"child"}},
				{Name: "parent", ParentName: "", ElementPath: []string{"parent"}},
			}
			sorted, err := SortDependencies(deps)
			Expect(err).ToNot(HaveOccurred())
			Expect(sorted).To(HaveLen(2))
			// Parent should come before child
			parentIdx := -1
			childIdx := -1
			for i, d := range sorted {
				if d.Name == "parent" {
					parentIdx = i
				}
				if d.Name == "child" {
					childIdx = i
				}
			}
			Expect(parentIdx).To(BeNumerically("<", childIdx))
		})

		It("detects cycles", func() {
			deps := []SpecSelectorDependency{
				{Name: "a", ParentName: "b", ElementPath: []string{"a"}},
				{Name: "b", ParentName: "a", ElementPath: []string{"b"}},
			}
			_, err := SortDependencies(deps)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cycle"))
		})

		It("handles same selector with different params", func() {
			deps := []SpecSelectorDependency{
				{Name: "containerMeta", ParentName: "container", ElementPath: []string{"metadata"}, SelectorParams: map[string]any{"name": "worker"}},
				{Name: "containerMeta", ParentName: "container", ElementPath: []string{"metadata"}, SelectorParams: map[string]any{"name": "main"}},
				{Name: "container", ParentName: "", ElementPath: []string{"spec", "containers"}},
			}
			sorted, err := SortDependencies(deps)
			Expect(err).ToNot(HaveOccurred())
			Expect(sorted).To(HaveLen(3))
			// container should come before both containerMeta entries
			containerIdx := -1
			for i, d := range sorted {
				if d.Name == "container" {
					containerIdx = i
				}
			}
			for i, d := range sorted {
				if d.Name == "containerMeta" {
					Expect(containerIdx).To(BeNumerically("<", i))
				}
			}
		})
	})

	Describe("CollectDependencies", func() {
		It("returns empty for rules without SET/DEFAULT enforcements", func() {
			reqEnforcement := policyApi.NewPolicyEnforcementRequired()
			reqEnforcement.SetRequired(true)
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector: "sel1",
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementRequired: reqEnforcement},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"sel1": *newSpecSelectorWithDependency(".", "parent", "child"),
			}
			deps := CollectDependencies(rules, specSelectors)
			Expect(deps).To(BeEmpty())
		})

		It("collects dependencies for SET enforcements", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector: "templateMeta",
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"templateMeta": *newSpecSelectorWithDependency(".spec.template.metadata", "", "spec.template.metadata"),
			}
			deps := CollectDependencies(rules, specSelectors)
			Expect(deps).To(HaveLen(1))
			Expect(deps[0].Name).To(Equal("templateMeta"))
			Expect(deps[0].ElementPath).To(Equal([]string{"spec", "template", "metadata"}))
		})

		It("collects dependencies for DEFAULT enforcements", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector: "nested",
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementDefault: &policyApi.PolicyEnforcementDefault{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"nested": *newSpecSelectorWithDependency(".spec.nested", "", "spec.nested"),
			}
			deps := CollectDependencies(rules, specSelectors)
			Expect(deps).To(HaveLen(1))
		})

		It("skips selectors without dependencies", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector: "root",
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"root": {Query: ptrString(".")},
			}
			deps := CollectDependencies(rules, specSelectors)
			Expect(deps).To(BeEmpty())
		})

		It("collects separate dependencies for same selector with different params", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector:   "containerMeta",
					SelectorParams: map[string]any{"containerName": "worker"},
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector:   "containerMeta",
					SelectorParams: map[string]any{"containerName": "main"},
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"containerMeta": *newSpecSelectorWithDependency(
					`.spec.containers[] | select(.name == $ARGS.named.containerName) | .metadata`,
					"containerByName",
					"metadata",
				),
				"containerByName": {Query: ptrString(`.spec.containers[] | select(.name == $ARGS.named.containerName)`)},
			}
			deps := CollectDependencies(rules, specSelectors)
			// Should collect TWO dependencies - one for each param variation
			Expect(deps).To(HaveLen(2))
			// Verify both param sets are captured
			params := []string{}
			for _, d := range deps {
				if name, ok := d.SelectorParams["containerName"].(string); ok {
					params = append(params, name)
				}
			}
			Expect(params).To(ContainElements("worker", "main"))
		})

		It("deduplicates same selector with same params", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector:   "containerMeta",
					SelectorParams: map[string]any{"containerName": "worker"},
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector:   "containerMeta",
					SelectorParams: map[string]any{"containerName": "worker"}, // same params
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementDefault: &policyApi.PolicyEnforcementDefault{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"containerMeta": *newSpecSelectorWithDependency(
					`.spec.containers[] | select(.name == $ARGS.named.containerName) | .metadata`,
					"containerByName",
					"metadata",
				),
				"containerByName": {Query: ptrString(`.spec.containers[] | select(.name == $ARGS.named.containerName)`)},
			}
			deps := CollectDependencies(rules, specSelectors)
			// Should collect only ONE dependency - same selector with same params
			Expect(deps).To(HaveLen(1))
		})

		It("stores selector params in dependency", func() {
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector:   "serviceMeta",
					SelectorParams: map[string]any{"serviceName": "api-server"},
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"serviceMeta": *newSpecSelectorWithDependency(
					`.spec.services[] | select(.name == $ARGS.named.serviceName) | .metadata`,
					"serviceByName",
					"metadata",
				),
				"serviceByName": {Query: ptrString(`.spec.services[] | select(.name == $ARGS.named.serviceName)`)},
			}
			deps := CollectDependencies(rules, specSelectors)
			Expect(deps).To(HaveLen(1))
			Expect(deps[0].SelectorParams).To(HaveKeyWithValue("serviceName", "api-server"))
		})

		It("recursively collects entire dependency chain (X->Y->Z)", func() {
			// Rule uses X, X depends on Y, Y depends on Z
			// All three should be collected
			rules := []ResolvedRule{
				{rule: &policyApi.EffectiveRulesEffectiveRulesInner{
					SpecSelector: "leafSelector",
					Enforcements: []policyApi.PolicyRuleEnforcement{
						{PolicyEnforcementSet: &policyApi.PolicyEnforcementSet{}},
					},
				}},
			}
			specSelectors := map[string]policyApi.BreakdownSpecSelector{
				"leafSelector": *newSpecSelectorWithDependency(
					".spec.level1.level2.leaf",
					"midSelector",
					"leaf",
				),
				"midSelector": *newSpecSelectorWithDependency(
					".spec.level1.level2",
					"rootSelector",
					"level2",
				),
				"rootSelector": *newSpecSelectorWithDependency(
					".spec.level1",
					"",
					"spec.level1",
				),
			}
			deps := CollectDependencies(rules, specSelectors)
			// Should collect all 3: rootSelector, midSelector, leafSelector
			Expect(deps).To(HaveLen(3))

			// Verify all three are present
			names := []string{}
			for _, d := range deps {
				names = append(names, d.Name)
			}
			Expect(names).To(ContainElements("leafSelector", "midSelector", "rootSelector"))
		})
	})

	Describe("EnsureDependencyPaths", func() {
		It("creates path from root when no parent", func() {
			manifest := map[string]any{
				"existing": "value",
			}
			deps := []SpecSelectorDependency{
				{Name: "sel", ElementPath: []string{"spec", "template", "metadata"}},
			}
			EnsureDependencyPaths(manifest, deps)

			Expect(manifest).To(HaveKey("spec"))
			spec := manifest["spec"].(map[string]any)
			Expect(spec).To(HaveKey("template"))
			template := spec["template"].(map[string]any)
			Expect(template).To(HaveKey("metadata"))
		})
	})
})

func ptrString(s string) *string {
	return &s
}

func newSpecSelectorWithDependency(query, parentSelector, element string) *policyApi.BreakdownSpecSelector {
	sel := &policyApi.BreakdownSpecSelector{
		Query: ptrString(query),
	}
	dep := policyApi.SpecSelectorDependency{}
	if parentSelector != "" {
		dep.SetSpecSelector(parentSelector)
	}
	dep.SetPath(element)
	sel.SetDependency(dep)
	return sel
}
