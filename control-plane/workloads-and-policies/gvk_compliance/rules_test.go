/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/stretchr/testify/assert"
)

// policyApplyTestCase selects the input and the expected output file for a
// single policy execution test case. A single filename is used for both
// (testing/inputs/<fname>.yaml and testing/expected/<fname>.yaml). The
// expected file always captures the whole result — the workload after the
// policy ran plus any validation errors — so mutations performed by
// set/default enforcements are observable alongside the errors.
type policyApplyTestCase struct {
	fname string
	// maxErrors overrides the default error cap for this test case.
	// Zero means use testMaxErrors.
	maxErrors int
}

// policyApplyTestInput is the YAML structure of the input file. The GVK is
// derived from the workload manifest's apiVersion/kind (same keys a real k8s
// resource uses). The policy carries its own breakdown inline — this is not
// ideal, but keeps these tests self-contained so they don't depend on a
// running policy-service. Breakdowns should mirror what the policy-service
// would seed for the same workload type (see
// control-plane/policy-service/pkg/services/system_mapping_*.go).
type policyApplyTestInput struct {
	Policy   *policyApi.GetGvkPolicyResp `json:"policy"`
	Workload map[string]interface{}      `json:"workload"`
}

// policyApplyTestResult is the YAML structure written to / read from the
// expected file. It captures whatever the policy execution produced so both
// mutations and validation errors are observable. The policy is also echoed
// back so the expected file is self-contained and reviewers don't have to
// flip back to the input file.
type policyApplyTestResult struct {
	Policy      *policyView            `json:"policy,omitempty"`
	Workload    map[string]interface{} `json:"workload,omitempty"`
	Errors      []testErrorSnapshot    `json:"errors,omitempty"`
	BrokenRules []brokenRuleSnapshot   `json:"brokenRules,omitempty"`
}

// policyView is a stripped-down projection of the input policy — only the
// workload type and the rules that drive the engine. The breakdown is not
// included; it is mapping data, available in the input file when needed for
// debugging.
type policyView struct {
	WorkloadType   policyApi.GVK                                 `json:"workloadType"`
	EffectiveRules []policyApi.EffectiveRulesEffectiveRulesInner `json:"effectiveRules,omitempty"`
}

type testErrorSnapshot struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// brokenRuleSnapshot captures a single broken-rule event (rule was skipped
// due to a mapping / formatter issue) so it can be asserted against the
// expected yaml instead of being lost to the logger.
type brokenRuleSnapshot struct {
	Field        string `json:"field"`
	SpecSelector string `json:"specSelector"`
	Message      string `json:"message"`
}

// testMaxErrors matches the default used by the submission path.
const testMaxErrors = 100

var policyApplyTests = []TableEntry{
	// image fails options, command fails minLen -> two accumulated errors.
	Entry("image and command violations", policyApplyTestCase{
		fname: "image_and_command_violations",
	}),
	// Same policy with compliant values -> no errors.
	Entry("image and command ok", policyApplyTestCase{
		fname: "image_and_command_ok",
	}),
	// default fills in missing image only, set always enforces
	// allowPrivilegeEscalation=false. Expected: container-1 keeps image=redhat
	// and gets allowPrivilegeEscalation=false; container-2 gets image=ubuntu
	// and allowPrivilegeEscalation flipped from true to false.
	Entry("set and default image and security context", policyApplyTestCase{
		fname: "set_and_default_image_and_sec_context",
	}),
	// Exercises selector discrimination across three spec-selector shapes
	// (entireWorkload, allContainers, appContainerByName) and multiple item
	// selectors on a single field:
	//   * Field "name" has two variations in the breakdown (entireWorkload vs
	//     allContainers) with different specRefs.
	//   * Field "image" uses the parameterized spec selector appContainerByName
	//     (selectorParams carries containerName) so only a specific container
	//     is validated.
	//   * Field "environmentVariables-Name" has one variation with multiple
	//     item selectors (allEnvVars and envVarFromConfigMap), driven by two
	//     separate rules.
	// Expected: 5 errors — one per rule (hits the maxErrors cap).
	Entry("spec and item selectors violations", policyApplyTestCase{
		fname: "spec_and_item_selectors_violations",
	}),
	// Same breakdown + rules, with a workload that satisfies every rule.
	Entry("spec and item selectors ok", policyApplyTestCase{
		fname: "spec_and_item_selectors_ok",
	}),
	// Exercises whole-instance set on itemized fields and a follow-up
	// sub-field set over multiple containers:
	//   * environmentVariables-FromConfigMap (custom formatter with an
	//     arrayKey of "name") is used to build entire env-var entries from
	//     several $ARGS.named values, with per-container targeting via
	//     appContainerByName.
	//   * environmentVariables-ConfigMapName (KeyValue formatter, scoped
	//     to the envVarFromConfigMap item selector) then overrides the
	//     configMap reference name on every env var added above, across
	//     all containers.
	// worker-3 matches no rule and stays bare, demonstrating that rules
	// don't touch unrelated containers.
	Entry("set envvars from configmap", policyApplyTestCase{
		fname: "set_envvars_from_configmap",
	}),
	// Sweep across every validation enforcement (max, min, step, maxLen, minLen,
	// options, match, maxQuantity, minQuantity, quantityStep, maxElements,
	// minElements, required, maxItems, minItems) against a synthetic
	// breakdown. Each rule's value is picked to fail, so the expected output
	// contains one error per rule — 15 in total, which needs maxErrors to be
	// bumped above the default submission cap of 5.
	Entry("get actions all", policyApplyTestCase{
		fname:     "get_enforcements_all",
		maxErrors: 30,
	}),
	// Sweep across the mutation side: scalar set, scalar default for a
	// present field (no-op) and an absent field (fills in), a nested set
	// that has to create intermediate maps, and whole-instance set via a
	// custom formatter with arrayKey merging across multiple containers.
	// Expected: 0 errors and a workload rewritten according to the rules.
	Entry("set actions all", policyApplyTestCase{
		fname: "set_actions_all",
	}),
	// Exercises the two patterns for operating on a map-valued field
	// (here: synthetic container-level "labels"):
	//   * a custom formatter that takes a key + value to add or override a
	//     single entry. Used by the "labels" field both to add a/b/c to
	//     every container (via allContainers) and to flip labels[b] on
	//     container-2 only (via appContainerByName).
	//   * KeyValueMap via the "labels-value" field with an allLabels item
	//     selector, which fans a new value across every existing key of the
	//     target map by relying on the CopyTargetKeys workaround in
	//     SetEnforcement to wire the formatter's generic placeholder into the
	//     target's actual keys.
	Entry("kvmap labels per container", policyApplyTestCase{
		fname: "kvmap_labels_per_container",
	}),

	// Exercises KeyValueMap with a nested specRef where intermediate objects
	// don't exist. The spec selector emits .spec (replicas only), and specRef
	// navigates to layer1.layer2 (neither exists). Expected: the engine
	// creates layer1 and layer2 to store the new key-value pair.
	Entry("kvmap with nested specRef", policyApplyTestCase{
		fname: "kvmap_with_nested_spec_ref",
	}),
	// Exercises KeyValueMap with specRef using the explicit { key, value }
	// shape: the JQ expression resolves $ARGS.named.key to "new-label" and
	// inserts the entry at .spec.selector.matchLabels[new-label]. The
	// CopyTargetKeys path is NOT triggered (key arg is present), so this
	// case validates the non-fan-out branch of the wrapped set expression.
	Entry("kvmap with specRef and explicit key/value", policyApplyTestCase{
		fname: "kvmap_with_spec_ref",
	}),
	// Exercises KeyValueMap with specRef using the value-only set form
	// (no "key" arg). This triggers the CopyTargetKeys fan-out, which the
	// runtime must apply *inside* the specRef-nested map — not at srcValue
	// root. Pre-fix the runtime would look up srcValue["value"] (nil
	// because the set expression nests the placeholder under specRef) and
	// clobber the parent of the actual labels map with nil; this case
	// pins the corrected navigation.
	Entry("kvmap with specRef and value-only set", policyApplyTestCase{
		fname: "kvmap_with_spec_ref_value_only",
	}),

	// Exercises every logBrokenRule path surfaced by the engine — missing
	// field, unknown spec selector, item-selector name not found, unknown
	// builtin formatter, invalid JQ, formatter missing expressions, invalid
	// regex / quantity limits, enforcements reading the wrong value shape, etc.
	// Each rule is crafted to hit a single path so the captured
	// brokenRules list documents which code branch produces which message.
	Entry("broken rules", policyApplyTestCase{
		fname: "broken_rules",
	}),
	// Five failing rules with maxErrors set to 2 — demonstrates that the
	// engine stops enqueuing errors once the cap is reached. Expected: only
	// two errors in the result.
	Entry("max errors cap", policyApplyTestCase{
		fname:     "max_errors_cap",
		maxErrors: 2,
	}),
	// Policy with rules but no breakdown — resolveRules emits a warning and
	// returns no resolved rules. Expected: no errors, no broken-rule events,
	// workload unchanged.
	Entry("no breakdown", policyApplyTestCase{
		fname: "no_breakdown",
	}),
	// Custom formatter with arrayKey.argName="id"; the set enforcement omits the
	// "id" arg from its value map. Expected: one 400 error of the form
	// "field items requires key 'id' to be set".
	Entry("set missing arrayKey arg", policyApplyTestCase{
		fname: "set_missing_arraykey_arg",
	}),
	// required:true on a present field → no error (the "found" branch of
	// RequiredEnforcement). step:5 on a string value → one real validation error
	// from validation.ToFloat failing. Combines the two into one test.
	Entry("required present and step non-numeric", policyApplyTestCase{
		fname: "required_present_and_step_nonnumeric",
	}),
	// required:true over a selector that resolves to multiple instances. One
	// container has image, another does not — required must fail because every
	// selected instance must provide the field.
	Entry("required multi occurrence missing item", policyApplyTestCase{
		fname: "required_multi_occurrence_missing_item",
	}),
	// blocked: true over a multi-instance spec selector;
	// fails when any selected instance carries a value.
	Entry("blocked multi occurrence", policyApplyTestCase{
		fname: "blocked_multi_occurrence",
	}),
	// blocked combined with an item selector — forbids any port that declares
	// a protocol; expects one 400 for the protocol-bearing port.
	Entry("blocked item selector", policyApplyTestCase{
		fname: "blocked_item_selector",
	}),
	// blocked: false accepts present values.
	Entry("blocked false", policyApplyTestCase{
		fname: "blocked_false",
	}),
	// Every value-scoped enforcement (max, min, step, match, options,
	// maxQuantity, minQuantity, quantityStep) runs its onValue callback
	// through asSlice so a rule bound to a scalar field and a rule bound
	// to an array field share the same implementation. This case binds
	// each enforcement to an array-valued field where a later element fails,
	// proving asSlice iterates every element.
	Entry("actions on arrays", policyApplyTestCase{
		fname: "actions_on_arrays",
	}),
	// Exercises the two branches of parseQuantity: the quantity-aware
	// enforcements receive a value that is not a string (int 42) or a string
	// that is not a valid k8s quantity ("notAQuantity"). Six rules, six
	// 400 validation errors — three per branch across maxQuantity,
	// minQuantity and quantityStep.
	Entry("quantity value errors", policyApplyTestCase{
		fname:     "quantity_value_errors",
		maxErrors: 10,
	}),
	// Exercises spec selector dependencies: when a spec selector declares a
	// dependency, the engine creates the missing intermediate path before
	// SET enforcements run. Without this, SET to a missing path silently skips
	// because the JQ selector yields no results.
	Entry("selector dependency creates missing paths", policyApplyTestCase{
		fname: "selector_dependency",
	}),
	// Exercises spec selector dependencies with multiple rules targeting
	// the same selector but with different params. Each (selector, params)
	// pair should have its path created independently.
	Entry("selector dependency with multiple params", policyApplyTestCase{
		fname: "selector_dependency_multi_params",
	}),
	// Exercises recursive dependency chain: X depends on Y, Y depends on Z.
	// All intermediate paths must be created in the correct order.
	Entry("selector dependency chain", policyApplyTestCase{
		fname: "selector_dependency_chain",
	}),
}

var _ = Describe("policy.Apply", func() {
	DescribeTable("applies policy rules to a workload and matches the expected outcome",
		func(tc policyApplyTestCase) {
			runPolicyApplyTest(tc, ".")
		},
		policyApplyTests,
	)
})

// gvkFromWorkload derives a GVK from a workload manifest using the standard
// k8s keys: apiVersion (e.g. "apps/v1" or "v1") and kind.
func gvkFromWorkload(workload map[string]interface{}) (policyApi.GVK, error) {
	apiVersion, _ := workload["apiVersion"].(string)
	kind, _ := workload["kind"].(string)
	if apiVersion == "" || kind == "" {
		return policyApi.GVK{}, fmt.Errorf("workload must have non-empty apiVersion and kind (got apiVersion=%q, kind=%q)", apiVersion, kind)
	}

	group, version := "", apiVersion
	if slash := strings.Index(apiVersion, "/"); slash >= 0 {
		group, version = apiVersion[:slash], apiVersion[slash+1:]
	}

	return policyApi.GVK{Group: *policyApi.NewNullableString(convutils.Ref(group)), Kind: kind, Version: version}, nil
}

func runPolicyApplyTest(tc policyApplyTestCase, basePath string) {
	testBrief := fmt.Sprintf("fname: %s", tc.fname)

	// Load input.
	inputYaml, err := testutils.LoadInputFile(tc.fname, basePath)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", tc.fname))

	var input policyApplyTestInput
	err = yaml.Unmarshal(inputYaml, &input)
	Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", tc.fname))
	Expect(input.Policy).NotTo(BeNil(), "input must define a policy")

	gvk, gvkErr := gvkFromWorkload(input.Workload)
	Expect(gvkErr).To(BeNil(), "workload must have apiVersion and kind")

	// The policy carries its own breakdown (loaded from the YAML). Make sure the
	// policy's workload type matches the workload manifest's GVK so Apply
	// resolves the right policy for this workload.
	input.Policy.SetWorkloadType(gvk)

	const projectId = "p1"

	// Wire the mock policy client so Apply resolves the test's policy.
	mock := &policies.SimpleMockClientFactory{
		SimpleMockClient: policies.SimpleMockClient{
			GvkPolicyOfProject: map[policies.ProjectAndType]*policyApi.GetGvkPolicyResp{},
		},
	}
	mock.WithGvkPolicy(projectId, gvk, input.Policy)
	policies.SetMock(mock)

	// Capture broken-rule messages instead of just logging them so they can
	// be asserted against the expected yaml. Always reset the hook after the
	// test so subsequent runs aren't polluted.
	var capturedBroken []brokenRuleSnapshot

	BrokenRuleHook = func(rule *policyApi.EffectiveRulesEffectiveRulesInner, msg string) {
		capturedBroken = append(capturedBroken, brokenRuleSnapshot{
			Field:        rule.Field,
			SpecSelector: rule.SpecSelector,
			Message:      msg,
		})
	}
	defer func() { BrokenRuleHook = nil }()

	// Run the policy.
	maxErrors := tc.maxErrors
	if maxErrors == 0 {
		maxErrors = testMaxErrors
	}

	resolvedRules := ResolveRules(input.Policy)

	// Extract spec selectors from breakdown for dependency resolution
	var specSelectors map[string]policyApi.BreakdownSpecSelector
	if bd := input.Policy.Breakdown.Get(); bd != nil && bd.SpecSelectors != nil {
		specSelectors = *bd.SpecSelectors
	}

	apiErrs := RunRulesWithDependencies(input.Workload, resolvedRules, specSelectors, maxErrors, AllCycles)

	// Build the result snapshot. The policy view is echoed back so the
	// expected file shows what was applied next to the resulting workload.
	result := policyApplyTestResult{
		Policy: &policyView{
			WorkloadType:   gvk,
			EffectiveRules: input.Policy.EffectiveRules,
		},
		Workload:    input.Workload,
		BrokenRules: capturedBroken,
	}
	for _, e := range apiErrs {
		result.Errors = append(result.Errors, testErrorSnapshot{
			Code:    e.Code,
			Message: e.Message,
		})
	}

	actualYaml, err := yaml.Marshal(result)
	Expect(err).To(BeNil(), testBrief)

	comparePolicyApplyResult(testBrief, tc.fname, actualYaml, basePath)
}

func comparePolicyApplyResult(testBrief, expectedFname string, actualYaml []byte, basePath string) {
	// testutils.LoadExpectedFile creates the file on the first run via
	// os.WriteFile, which does not create parent directories. Make sure the
	// expected directory exists so the first run produces the expected file
	// with the "review it!" message instead of a confusing "no such file or
	// directory" error.
	expectedDir := fmt.Sprintf("%s/%s", basePath, testutils.DefaultExpectedDir)
	Expect(os.MkdirAll(expectedDir, 0600)).To(BeNil(), "failed to create expected directory")

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	if assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testBrief) {
		return
	}

	tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
	if osErr := os.WriteFile(tmpPath, actualYaml, 0755); osErr != nil {
		assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osErr))
		return
	}
	assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
}
