package system_mapping_test

// Test harness for Layer 3+4. Glue between the per-GVK skeleton, a hand-
// authored set-everything-by-scope policy YAML, the in-memory breakdown
// built by buildLayer3Breakdown, and the gvk_compliance engine.
//
// The harness implements the round-trip pipeline described at the top of
// system_mapping_layer3_test.go in three pieces:
//
//   loadLayer3PolicyFile  — read a single policy YAML and return its rules
//   composeLayer3Policy   — wrap rules + a runtime-built breakdown into a
//                           policyApi.GetGvkPolicyResp the engine resolves
//   runLayer3             — execute the engine over a (deep-copied) skeleton
//                           and return the rendered manifest plus any
//                           engine-side validation errors / broken rules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/gvk_compliance"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

// layer3PolicyFile is the schema used by every per-case policy.yaml under
// testing/inputs/<gvk>/cases/<case>/. Only effectiveRules are authored;
// the breakdown is built at runtime from the system mapping (so adding a
// new field to the mapping requires no policy YAML changes for unrelated
// fields, and policies stay focused on the rules they exercise).
type layer3PolicyFile struct {
	Description    string                                        `json:"description,omitempty"`
	EffectiveRules []policyApi.EffectiveRulesEffectiveRulesInner `json:"effectiveRules"`
}

// loadLayer3PolicyFile reads one policy YAML and returns its effective
// rules. Returns an error if the file cannot be opened or its content
// fails to parse as a layer3PolicyFile.
func loadLayer3PolicyFile(path string) ([]policyApi.EffectiveRulesEffectiveRulesInner, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc layer3PolicyFile
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return doc.EffectiveRules, nil
}

// composeLayer3Policy builds a policyApi.GetGvkPolicyResp from the runtime
// breakdown plus the rules loaded from disk. The output is what
// gvk_compliance.ResolveRules expects: a policy whose breakdown lists every
// (specSelector, fieldGroup, field) combination the mapping defines, and
// whose effectiveRules drive the SET cycle.
func composeLayer3Policy(
	gvk api.Gvk,
	rules []policyApi.EffectiveRulesEffectiveRulesInner,
) (*policyApi.GetGvkPolicyResp, error) {
	breakdownJSON, err := buildLayer3Breakdown(gvk)
	if err != nil {
		return nil, err
	}

	// JSON-roundtrip the breakdown into the typed policyApi struct: the
	// Go shape is generated from the same OpenAPI source as the API
	// payload, so an identical JSON tree maps cleanly across packages.
	rawBreakdown, err := json.Marshal(breakdownJSON)
	if err != nil {
		return nil, fmt.Errorf("marshal breakdown: %w", err)
	}
	var breakdown policyApi.BreakdownBreakdown
	if err := json.Unmarshal(rawBreakdown, &breakdown); err != nil {
		return nil, fmt.Errorf("unmarshal breakdown: %w", err)
	}

	policy := policyApi.GetGvkPolicyResp{
		Id:             "layer3-set-everything",
		Scope:          policyApi.SCOPETYPE_TENANT,
		EffectiveRules: rules,
	}
	policy.SetWorkloadType(policyApi.GVK{Group: *policyApi.NewNullableString(gvk.Group), Kind: gvk.Kind, Version: gvk.Version})
	policy.SetBreakdown(breakdown)

	return &policy, nil
}

// layer3RunResult bundles everything the harness produces for a single run:
// the post-policy workload manifest, any engine validation errors, and any
// broken-rule events captured during execution. Callers compare the
// manifest to the golden YAML and assert that errors / broken rules are
// empty (a successful SET-everything must not produce either).
type layer3RunResult struct {
	Workload    map[string]any
	Errors      []*policyApi.Error
	BrokenRules []layer3BrokenRule
}

// layer3BrokenRule mirrors the BrokenRuleHook payload but in a comparison-
// friendly shape. Anything captured here means a rule was logically dropped
// by the engine (typically a mapping inconsistency) — for SET-everything,
// the expected count is zero.
type layer3BrokenRule struct {
	Field        string
	SpecSelector string
	Message      string
}

// runLayer3 executes the policy engine over a deep copy of skeleton and
// returns the rendered manifest. The skeleton itself is not mutated, so
// the same loaded skeleton can drive multiple test runs.
func runLayer3(
	ctx context.Context,
	policy *policyApi.GetGvkPolicyResp,
	skeleton map[string]any,
	maxErrors int,
) layer3RunResult {
	workload := deepCopyMap(skeleton)

	var captured []layer3BrokenRule
	gvk_compliance.BrokenRuleHook = func(rule *policyApi.EffectiveRulesEffectiveRulesInner, msg string) {
		captured = append(captured, layer3BrokenRule{
			Field:        rule.Field,
			SpecSelector: rule.SpecSelector,
			Message:      msg,
		})
	}
	defer func() { gvk_compliance.BrokenRuleHook = nil }()

	resolved := gvk_compliance.ResolveRules(policy)

	// Pass spec selectors to enable dependency path creation
	var specSelectors map[string]policyApi.BreakdownSpecSelector
	if bd := policy.Breakdown.Get(); bd != nil && bd.SpecSelectors != nil {
		specSelectors = *bd.SpecSelectors
	}

	errs := gvk_compliance.RunRulesWithDependencies(workload, resolved, specSelectors, maxErrors, gvk_compliance.AllCycles)

	return layer3RunResult{
		Workload:    workload,
		Errors:      errs,
		BrokenRules: captured,
	}
}

// deepCopyMap clones a JSON-shaped map[string]any so callers can mutate the
// copy without affecting the input. Uses encoding/json, which is the same
// shape the engine itself relies on (any non-JSON-shaped values would not
// survive a round-trip through gojq either).
//
// Decoding uses UseNumber() so numbers become json.Number rather than float64,
// mirroring how the real service loads the workload spec (gojsonq.LoadJson).
// This keeps the target of set/default merges on the same numeric type as the
// policy values, which the engine canonicalizes to json.Number.
func deepCopyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	raw, err := json.Marshal(in)
	Expect(err).NotTo(HaveOccurred())
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var out map[string]any
	err = decoder.Decode(&out)
	Expect(err).NotTo(HaveOccurred())
	return out
}
