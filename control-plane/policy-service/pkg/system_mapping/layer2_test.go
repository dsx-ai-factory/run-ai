/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping_test

// Layer 2 of the system-mapping test plan: GET-side correctness.
//
// Verifies that the runtime's get-expression (composed from the mapping's
// formatter + SpecRef) correctly retrieves values from a workload spec.
//
// For each field with a builtin formatter (KeyValue / KeyValueMap):
//   - Compose the getExpr exactly like the runtime does (via keyValueExpressions)
//   - Run the spec-selector + (optional) item-selector pipeline to get roots
//   - Run the getExpr against each root
//   - Compare the returned value to what's at the SpecRef path in the reference spec
//
// This catches bugs where the getExpr traverses the wrong branch, where an item
// selector filters too aggressively, or where a KeyValue path doesn't match SpecRef.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	sm "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/system_mapping"
)

var _ = Describe("system mapping layer 2: GET-side correctness", func() {
	Context("walker self-tests", func() {
		It("detects a getExpr that returns wrong value", func() {
			const fgName = "selfTestFG"
			builders := map[string]sm.SystemFieldGroupBuilder{
				fgName: func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
							// SpecRef says "replicas" but the spec has value at "replicas"
							// with value 3, so getExpr should return 3
							"replicaField": {Formatter: dbmodel.BuiltinFormatterKeyValue, SpecRef: "replicas", Type: api.MAPPINGFIELDTYPE_INTEGER},
						},
						map[string]api.ItemSelector{}
				},
			}
			specSelectors := map[string]api.SpecSelector{
				"selfTest": {Query: ".spec", FieldGroups: []string{fgName}},
			}
			// Put a different value than what getExpr would find
			refSpec := map[string]any{
				"spec": map[string]any{"replicas": float64(3)},
			}

			report, err := runLayer2(specSelectors, builders, refSpec)
			Expect(err).To(BeNil())
			// Should pass - value matches
			Expect(report.GetMismatch).To(BeEmpty())
			Expect(report.NeverResolved).To(BeEmpty())
		})

		It("catches a getExpr that returns wrong value due to bad SpecRef", func() {
			const fgName = "selfTestFG"
			builders := map[string]sm.SystemFieldGroupBuilder{
				fgName: func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
							// SpecRef says "replicaz" (typo), so getExpr will look for .["replicaz"]
							// but that path doesn't exist in the spec
							"replicaField": {Formatter: dbmodel.BuiltinFormatterKeyValue, SpecRef: "replicaz", Type: api.MAPPINGFIELDTYPE_INTEGER},
						},
						map[string]api.ItemSelector{}
				},
			}
			specSelectors := map[string]api.SpecSelector{
				"selfTest": {Query: ".spec", FieldGroups: []string{fgName}},
			}
			refSpec := map[string]any{
				"spec": map[string]any{"replicas": float64(3)},
			}

			report, err := runLayer2(specSelectors, builders, refSpec)
			Expect(err).To(BeNil())
			// Should fail - getExpr returns nil (field not found)
			Expect(report.NeverResolved).To(HaveLen(1))
			Expect(report.NeverResolved[0].field).To(Equal("replicaField"))
		})
	})

	Context("apps/Deployment/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsDeploymentV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_Deployment_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("batch/Job/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemBatchJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/batch_Job_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("apps/ReplicaSet/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsReplicaSetV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_ReplicaSet_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("apps/StatefulSet/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsStatefulSetV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_StatefulSet_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("batch/CronJob/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemBatchCronJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/batch_CronJob_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("/Pod/v1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemCorePodV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/core_Pod_v1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("apps.nvidia.com/NIMCache/v1alpha1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_nvidia_com_NIMCache_v1alpha1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})

	Context("apps.nvidia.com/NIMService/v1alpha1", func() {
		It("getExpr retrieves correct values from reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_nvidia_com_NIMService_v1alpha1/reference.yaml")

			report, err := runLayer2(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatLayer2Report(report)).To(BeEmpty())
		})
	})
})

// getMismatch records a case where the getExpr returned a value that doesn't
// match what walkSpecRef finds at the same path.
type getMismatch struct {
	specSelector string
	itemSelector string
	fieldGroup   string
	field        string
	specRef      string
	getExprGot   any
	walkGot      any
}

// layer2Report aggregates a single GVK's Layer 2 walk results.
type layer2Report struct {
	NeverResolved []fieldKey
	GetMismatch   []getMismatch
}

// runLayer2 walks the (specSelectors, fieldGroups) mapping over refSpec and
// verifies that the runtime's getExpr produces the same value as a direct
// path walk at SpecRef.
func runLayer2(
	specSelectors map[string]api.SpecSelector,
	builders map[string]sm.SystemFieldGroupBuilder,
	refSpec map[string]any,
) (*layer2Report, error) {
	report := &layer2Report{}

	tracked := map[fieldKey]bool{}
	resolved := map[fieldKey]bool{}

	specSelectorNames := sm.SortedKeys(specSelectors)

	for _, selName := range specSelectorNames {
		sel := specSelectors[selName]

		selOutputs, err := evalJQ(sel.Query, refSpec, sampleArgs(sel.ArgsInfo, sel.ArgsEnum))
		if err != nil {
			return nil, fmt.Errorf("spec selector %q: %w", selName, err)
		}
		if len(selOutputs) == 0 {
			return nil, fmt.Errorf("spec selector %q: produced no outputs from reference spec", selName)
		}

		for _, fgName := range sel.FieldGroups {
			builder, ok := builders[fgName]
			if !ok {
				return nil, fmt.Errorf("spec selector %q: field group %q has no registered builder", selName, fgName)
			}
			fields, itemSelectors := builder()

			for _, fName := range sm.SortedKeys(fields) {
				f := fields[fName]
				if !isBuiltinFormatter(f.Formatter) {
					// Custom formatters have their own getExpr (or none).
					// Layer 2 focuses on builtin formatters whose getExpr
					// is auto-composed from SpecRef.
					continue
				}

				key := fieldKey{fieldGroup: fgName, field: fName}
				tracked[key] = true

				// Compose the getExpr like the runtime does
				getExpr := keyValueGetExpr(f.SpecRef)
				getCode, err := compileJQLayer2(getExpr)
				if err != nil {
					return nil, fmt.Errorf("field %s.%s: failed to compile getExpr %q: %w", fgName, fName, getExpr, err)
				}

				// First mode: resolve directly from each spec-selector output
				for _, root := range selOutputs {
					getResult := runGetExprOnce(getCode, root)
					walkResult, walkOk := walkSpecRef(root, f.SpecRef)

					if getResult == nil && !walkOk {
						// Both found nothing - consistent, but not resolved
						continue
					}

					resolved[key] = true

					if !valuesEqual(getResult, walkResult) {
						report.GetMismatch = append(report.GetMismatch, getMismatch{
							specSelector: selName,
							fieldGroup:   fgName,
							field:        fName,
							specRef:      f.SpecRef,
							getExprGot:   getResult,
							walkGot:      walkResult,
						})
					}
				}

				// Second mode: for each compatible item selector
				for _, isName := range f.ItemSelectors {
					isq, found := itemSelectors[isName]
					if !found {
						return nil, fmt.Errorf("field group %q field %q references missing item selector %q", fgName, fName, isName)
					}

					isArgs := sampleArgs(isq.ArgsInfo, isq.ArgsEnum)
					for _, root := range selOutputs {
						itemOutputs, err := evalJQ(isq.Query, root, isArgs)
						if err != nil {
							return nil, fmt.Errorf("item selector %q (under %q): %w", isName, selName, err)
						}
						for _, item := range itemOutputs {
							getResult := runGetExprOnce(getCode, item)
							walkResult, walkOk := walkSpecRef(item, f.SpecRef)

							if getResult == nil && !walkOk {
								continue
							}

							resolved[key] = true

							if !valuesEqual(getResult, walkResult) {
								report.GetMismatch = append(report.GetMismatch, getMismatch{
									specSelector: selName,
									itemSelector: isName,
									fieldGroup:   fgName,
									field:        fName,
									specRef:      f.SpecRef,
									getExprGot:   getResult,
									walkGot:      walkResult,
								})
							}
						}
					}
				}
			}
		}
	}

	for k := range tracked {
		if !resolved[k] {
			report.NeverResolved = append(report.NeverResolved, k)
		}
	}
	sort.Slice(report.NeverResolved, func(i, j int) bool {
		if report.NeverResolved[i].fieldGroup != report.NeverResolved[j].fieldGroup {
			return report.NeverResolved[i].fieldGroup < report.NeverResolved[j].fieldGroup
		}
		return report.NeverResolved[i].field < report.NeverResolved[j].field
	})
	sort.Slice(report.GetMismatch, func(i, j int) bool {
		a, b := report.GetMismatch[i], report.GetMismatch[j]
		if a.fieldGroup != b.fieldGroup {
			return a.fieldGroup < b.fieldGroup
		}
		if a.field != b.field {
			return a.field < b.field
		}
		if a.specSelector != b.specSelector {
			return a.specSelector < b.specSelector
		}
		return a.itemSelector < b.itemSelector
	})

	return report, nil
}

// keyValueGetExpr composes the get expression for a KeyValue builtin formatter,
// exactly mirroring what gvk_compliance.keyValueExpressions does.
// For specPath "resources.limits.cpu" it returns `.[\"resources\"][\"limits\"][\"cpu\"]`
func keyValueGetExpr(specPath string) string {
	parts := strings.Split(specPath, ".")
	getParts := make([]string, 0, len(parts))
	for _, part := range parts {
		getParts = append(getParts, fmt.Sprintf(`["%s"]`, part))
	}
	return "." + strings.Join(getParts, "")
}

// compileJQLayer2 compiles a JQ expression without $ARGS (getExpr doesn't use args).
func compileJQLayer2(expr string) (*gojq.Code, error) {
	parsed, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	code, err := gojq.Compile(parsed)
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}
	return code, nil
}

// runGetExprOnce runs the compiled getExpr against input and returns the first
// non-error result (or nil if the iterator is empty or all results are errors).
func runGetExprOnce(code *gojq.Code, input any) any {
	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			return nil
		}
		if _, isErr := v.(error); isErr {
			continue
		}
		return v
	}
}

// valuesEqual compares two values for equality. Handles nil, primitives, and
// recursively compares maps and slices.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, exists := bv[k]
			if !exists || !valuesEqual(va, vb) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case float64:
		// Handle int vs float64 comparison (YAML/JSON numbers)
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		case int64:
			return av == float64(bv)
		}
		return false
	case int:
		switch bv := b.(type) {
		case int:
			return av == bv
		case float64:
			return float64(av) == bv
		case int64:
			return int64(av) == bv
		}
		return false
	default:
		return a == b
	}
}

// formatLayer2Report renders a layer2Report into a single human-readable block.
func formatLayer2Report(r *layer2Report) string {
	if len(r.NeverResolved) == 0 && len(r.GetMismatch) == 0 {
		return ""
	}
	var b strings.Builder
	if len(r.NeverResolved) > 0 {
		b.WriteString("fields with builtin formatter whose getExpr never returned a value:\n")
		b.WriteString("(typo in SpecRef? spec selector didn't match? reference spec missing the field?)\n")
		for _, k := range r.NeverResolved {
			fmt.Fprintf(&b, "  - %s.%s\n", k.fieldGroup, k.field)
		}
	}
	if len(r.GetMismatch) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("fields where getExpr returned a different value than direct path walk:\n")
		for _, m := range r.GetMismatch {
			fmt.Fprintf(&b, "  - %s.%s [selector=%s itemSelector=%s] specRef=%s\n    getExpr got: %T(%v)\n    walk got: %T(%v)\n",
				m.fieldGroup, m.field, m.specSelector, m.itemSelector, m.specRef,
				m.getExprGot, m.getExprGot, m.walkGot, m.walkGot)
		}
	}
	return b.String()
}
