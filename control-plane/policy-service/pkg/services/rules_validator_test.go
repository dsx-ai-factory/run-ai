/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("RulesValidator", func() {
	pa := &PolicyApiService{}

	builtinName := "KeyValue"
	customSetExpr := `{"x": $ARGS.named.val}`
	customGetExpr := `.x`

	builtinFormatter := api.BreakdownFormatter{Builtin: &builtinName}
	customFormatter := api.BreakdownFormatter{
		SetExpr:  &customSetExpr,
		GetExpr:  &customGetExpr,
		ArgsInfo: map[string]api.MappingFieldType{"val": api.MAPPINGFIELDTYPE_STRING},
	}
	customIntArgFormatter := api.BreakdownFormatter{
		SetExpr:  &customSetExpr,
		GetExpr:  &customGetExpr,
		ArgsInfo: map[string]api.MappingFieldType{"val": api.MAPPINGFIELDTYPE_INTEGER},
	}
	noGetFormatter := api.BreakdownFormatter{
		SetExpr: &customSetExpr,
	}

	specSelQuery := `.spec.containers[] | select(.name == $ARGS.named.containerName)`

	mkItemSelectors := func(queries map[string]string) map[string]api.BreakdownItemSelector {
		if queries == nil {
			return nil
		}
		result := make(map[string]api.BreakdownItemSelector, len(queries))
		for name, query := range queries {
			result[name] = api.BreakdownItemSelector{Query: query}
		}
		return result
	}

	mkBreakdown := func(fieldName string, fieldType api.MappingFieldType, formatter api.BreakdownFormatter, itemSelectors map[string]string) *api.BreakdownBreakdown {
		return &api.BreakdownBreakdown{
			Fields: map[string][]api.BreakdownFieldVariation{
				fieldName: {{
					SpecSelectors: []string{"sel1"},
					Type:          fieldType,
					Formatter:     formatter,
					ItemSelectors: mkItemSelectors(itemSelectors),
				}},
			},
			SpecSelectors: map[string]api.BreakdownSpecSelector{
				"sel1": {
					Query:    specSelQuery,
					ArgsInfo: map[string]api.MappingFieldType{"containerName": api.MAPPINGFIELDTYPE_STRING},
					ArgsEnum: map[string][]string{},
				},
			},
		}
	}

	mkRule := func(field, specSel string, enforcement api.PolicyRuleEnforcement) []api.PolicyRule {
		return []api.PolicyRule{{
			Field:          field,
			SpecSelector:   specSel,
			SelectorParams: map[string]interface{}{"containerName": "worker-1"},
			Enforcements:   []api.PolicyRuleEnforcement{enforcement},
		}}
	}

	mkRuleWithEnforcements := func(field, specSel string, enforcements ...api.PolicyRuleEnforcement) []api.PolicyRule {
		return []api.PolicyRule{{
			Field:          field,
			SpecSelector:   specSel,
			SelectorParams: map[string]interface{}{"containerName": "worker-1"},
			Enforcements:   enforcements,
		}}
	}

	Context("nil breakdown", func() {
		It("returns nil when breakdown is nil", func() {
			rules := []api.PolicyRule{{Field: "f", SpecSelector: "s", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}}}
			Expect(pa.validateRules(nil, nil, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	Context("field lookup", func() {
		bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects unknown field", func() {
			rules := mkRule("unknownField", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`field "unknownField" not found under sel1`))
		})

		It("accepts known field", func() {
			rules := mkRule("fieldA", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	Context("spec selector lookup", func() {
		bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects unknown spec selector for a field", func() {
			rules := []api.PolicyRule{{Field: "fieldA", SpecSelector: "badSel", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not found under specSelector"))
			Expect(err.Details).To(ContainSubstring("sel1"))
		})
	})

	Context("item selector", func() {
		bdWithItems := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, map[string]string{
			"iSel1": `(.items // [])[]`,
			"iSel2": `(.items // [])[] | select(.name == $ARGS.named.itemName)`,
		})
		bdNoItems := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects itemSelector when field has none", func() {
			rules := []api.PolicyRule{{
				Field: "fieldA", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSel1"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bdNoItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("does not accept an itemSelector"))
		})

		It("rejects unknown itemSelector", func() {
			rules := []api.PolicyRule{{
				Field: "fieldA", SpecSelector: "sel1", ItemSelector: convutils.Ref("badISel"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bdWithItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not valid for this field"))
			Expect(err.Details).To(ContainSubstring("iSel1"))
		})

		It("requires itemSelector when field has them", func() {
			rules := []api.PolicyRule{{
				Field: "fieldA", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bdWithItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("itemSelector is required"))
		})

		It("accepts valid itemSelector", func() {
			rules := []api.PolicyRule{{
				Field: "fieldA", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSel1"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			Expect(pa.validateRules(nil, bdWithItems, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	Context("duplicate rules", func() {
		bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects duplicate field+selector+params combination", func() {
			enforcement := api.PolicyRuleEnforcement{Required: convutils.Ref(true)}
			rules := []api.PolicyRule{
				{Field: "fieldA", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{enforcement}},
				{Field: "fieldA", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{enforcement}},
			}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("duplicate rule"))
		})
	})

	Context("selector params", func() {
		bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects unknown selector param", func() {
			rules := []api.PolicyRule{{
				Field: "fieldA", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"unknownArg": "x"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not a recognized argument"))
		})

		It("accepts valid selector param", func() {
			rules := mkRule("fieldA", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	Context("single enforcement validation", func() {
		bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("rejects enforcement with no fields set", func() {
			rules := mkRule("fieldA", "sel1", api.PolicyRuleEnforcement{})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("exactly one non-nil field"))
		})

		It("rejects enforcement with multiple fields set", func() {
			rules := mkRule("fieldA", "sel1", api.PolicyRuleEnforcement{
				Required: convutils.Ref(true),
				Length:   &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1))},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("exactly one non-nil field"))
		})
	})

	Context("enforcement type compatibility", func() {
		type entry struct {
			fieldType   api.MappingFieldType
			enforcement api.PolicyRuleEnforcement
			shouldError bool
			desc        string
		}

		DescribeTable("validates enforcement-type combinations",
			func(e entry) {
				bd := mkBreakdown("fld", e.fieldType, builtinFormatter, nil)
				rules := mkRule("fld", "sel1", e.enforcement)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				if e.shouldError {
					Expect(err).NotTo(BeNil(), "expected error for %s on %s", e.desc, e.fieldType)
					Expect(err.Message).To(ContainSubstring(e.desc))
				} else {
					Expect(err).To(BeNil(), "unexpected error: %v", err)
				}
			},
			// set/default/required: allowed for all types
			Entry("set on string", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "v"}}, false, ""}),
			Entry("required on integer", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{Required: convutils.Ref(true)}, false, ""}),
			Entry("required on object", entry{api.MAPPINGFIELDTYPE_OBJECT, api.PolicyRuleEnforcement{Required: convutils.Ref(true)}, false, ""}),

			// string enforcements (allowed on string, namesArray, strings)
			Entry("match on string", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Match: convutils.Ref("^abc")}, false, ""}),
			Entry("match on namesArray", entry{api.MAPPINGFIELDTYPE_NAMES_ARRAY, api.PolicyRuleEnforcement{Match: convutils.Ref(".*")}, false, ""}),
			Entry("match on strings", entry{api.MAPPINGFIELDTYPE_STRINGS, api.PolicyRuleEnforcement{Match: convutils.Ref(".*")}, false, ""}),
			Entry("match on integer REJECTED", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{Match: convutils.Ref(".*")}, true, "match"}),
			Entry("match on boolean REJECTED", entry{api.MAPPINGFIELDTYPE_BOOLEAN, api.PolicyRuleEnforcement{Match: convutils.Ref(".*")}, true, "match"}),
			Entry("options on string", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Options: &[]string{"a", "b"}}, false, ""}),
			Entry("options on strings", entry{api.MAPPINGFIELDTYPE_STRINGS, api.PolicyRuleEnforcement{Options: &[]string{"a"}}, false, ""}),
			Entry("options on quantity REJECTED", entry{api.MAPPINGFIELDTYPE_QUANTITY, api.PolicyRuleEnforcement{Options: &[]string{"a"}}, true, "options"}),
			Entry("length(min) on string", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1))}}, false, ""}),
			Entry("length(min) on strings", entry{api.MAPPINGFIELDTYPE_STRINGS, api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1))}}, false, ""}),
			Entry("length(min) on integer REJECTED", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1))}}, true, "length"}),
			Entry("length(max) on string", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Max: convutils.Ref(int32(10))}}, false, ""}),
			Entry("length(max) on strings", entry{api.MAPPINGFIELDTYPE_STRINGS, api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Max: convutils.Ref(int32(10))}}, false, ""}),

			// numeric enforcements (allowed on integer, number, integers, numbers)
			Entry("range(min) on integer", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}}, false, ""}),
			Entry("range(min) on number", entry{api.MAPPINGFIELDTYPE_NUMBER, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1.5))}}, false, ""}),
			Entry("range(min) on integers", entry{api.MAPPINGFIELDTYPE_INTEGERS, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}}, false, ""}),
			Entry("range(min) on numbers", entry{api.MAPPINGFIELDTYPE_NUMBERS, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1.5))}}, false, ""}),
			Entry("range(min) on string REJECTED", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}}, true, "range"}),
			Entry("range(min) on quantity REJECTED", entry{api.MAPPINGFIELDTYPE_QUANTITY, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}}, true, "range"}),
			Entry("range(max) on integer", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Max: convutils.Ref(float32(100))}}, false, ""}),
			Entry("range(max) on integers", entry{api.MAPPINGFIELDTYPE_INTEGERS, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Max: convutils.Ref(float32(100))}}, false, ""}),
			Entry("range(step) on number", entry{api.MAPPINGFIELDTYPE_NUMBER, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Step: convutils.Ref(float32(5))}}, false, ""}),
			Entry("range(step) on numbers", entry{api.MAPPINGFIELDTYPE_NUMBERS, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Step: convutils.Ref(float32(5))}}, false, ""}),
			Entry("range(step) on boolean REJECTED", entry{api.MAPPINGFIELDTYPE_BOOLEAN, api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Step: convutils.Ref(float32(5))}}, true, "range"}),

			// quantity enforcements
			Entry("quantityRange(min) on quantity", entry{api.MAPPINGFIELDTYPE_QUANTITY, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("100Mi")}}, false, ""}),
			Entry("quantityRange(max) on quantity", entry{api.MAPPINGFIELDTYPE_QUANTITY, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Max: convutils.Ref("8Gi")}}, false, ""}),
			Entry("quantityRange(step) on quantity", entry{api.MAPPINGFIELDTYPE_QUANTITY, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Step: convutils.Ref("256Mi")}}, false, ""}),
			Entry("quantityRange on integer REJECTED", entry{api.MAPPINGFIELDTYPE_INTEGER, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("100Mi")}}, true, "quantityRange"}),
			Entry("quantityRange on string REJECTED", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("100Mi")}}, true, "quantityRange"}),
			Entry("quantityRange on object REJECTED", entry{api.MAPPINGFIELDTYPE_OBJECT, api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("100Mi")}}, true, "quantityRange"}),

			// array enforcements
			Entry("elements(min) on strings", entry{api.MAPPINGFIELDTYPE_STRINGS, api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(1))}}, false, ""}),
			Entry("elements(min) on integers", entry{api.MAPPINGFIELDTYPE_INTEGERS, api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(1))}}, false, ""}),
			Entry("elements(max) on numbers", entry{api.MAPPINGFIELDTYPE_NUMBERS, api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Max: convutils.Ref(int32(10))}}, false, ""}),
			Entry("elements on string REJECTED", entry{api.MAPPINGFIELDTYPE_STRING, api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(1))}}, true, "elements"}),
			Entry("elements on boolean REJECTED", entry{api.MAPPINGFIELDTYPE_BOOLEAN, api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(1))}}, true, "elements"}),
		)
	})

	Context("items requires itemSelectors", func() {
		bdWithItems := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, map[string]string{
			"iSel": `(.items // [])[]`,
		})
		bdNoItems := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("allows items when field has itemSelectors", func() {
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSel"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements: []api.PolicyRuleEnforcement{{
					Items: &api.PolicyEnforcementItemsItems{Min: convutils.Ref(int32(1))},
				}},
			}}
			Expect(pa.validateRules(nil, bdWithItems, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects items(min) when field has no itemSelectors", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Items: &api.PolicyEnforcementItemsItems{Min: convutils.Ref(int32(1))},
			})
			err := pa.validateRules(nil, bdNoItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("items is only applicable"))
		})

		It("rejects items(max) when field has no itemSelectors", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Items: &api.PolicyEnforcementItemsItems{Max: convutils.Ref(int32(5))},
			})
			err := pa.validateRules(nil, bdNoItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("items is only applicable"))
		})
	})

	Context("int-range (length/elements/items) consistency validation", func() {
		bdString := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
		bdArray := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRINGS, builtinFormatter, nil)
		bdWithItems := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, map[string]string{
			"iSel": `(.items // [])[]`,
		})

		It("rejects length with no bounds", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{}})
			err := pa.validateRules(nil, bdString, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("at least one of min/max"))
		})

		It("rejects length where min > max", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(10)), Max: convutils.Ref(int32(5))},
			})
			err := pa.validateRules(nil, bdString, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must be less than or equal to"))
		})

		It("rejects elements with negative bound", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(-1))},
			})
			err := pa.validateRules(nil, bdArray, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must be non-negative"))
		})

		It("rejects items with no bounds", func() {
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSel"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Items: &api.PolicyEnforcementItemsItems{}}},
			}}
			err := pa.validateRules(nil, bdWithItems, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("at least one of min/max"))
		})
	})

	Context("KeyValue set/default validation", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("accepts set with single 'value' key", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "hello"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects set with wrong key", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"bad": "hello"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must use key"))
		})

		It("rejects set with multiple keys", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "a", "extra": "b"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("exactly one entry"))
		})

		It("rejects set with wrong value type for integer field", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "notAnInt"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected integer"))
		})
	})

	Context("custom formatter set/default validation", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customFormatter, nil)

		It("accepts set with valid arg from setExpr", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": "hello"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects set with unknown arg", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"badArg": "hello"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not a recognized argument"))
			Expect(err.Details).To(ContainSubstring("val"))
		})

		It("rejects set integer parameter not in peer options", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customIntArgFormatter, nil)
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"1", "2"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": 5}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with options"))
			Expect(err.Message).To(ContainSubstring(`must be one of [1, 2]`))
		})

		It("rejects set integer parameter that fails peer match", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customIntArgFormatter, nil)
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Match: convutils.Ref("^[12]$")},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": 5}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("does not match pattern"))
		})
	})

	Context("quantityRange value validation", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_QUANTITY, builtinFormatter, nil)

		It("accepts valid quantities", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("10Mi"),
					Max:  convutils.Ref("1Gi"),
					Step: convutils.Ref("256Mi"),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects invalid quantity string", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("notAQuantity")},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not a valid Kubernetes resource quantity"))
		})

		It("rejects quantityRange where min > max", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min: convutils.Ref("2Gi"),
					Max: convutils.Ref("1Gi"),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must be less than or equal to"))
		})

		It("rejects empty quantityRange", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("at least one of min/max/step"))
		})

		It("rejects quantityRange where step exceeds max-min", func() {
			// min=100Mi, max=1Gi -> span ~= 924Mi; step=2Gi leaves only `min`
			// as the single valid quantity.
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("100Mi"),
					Max:  convutils.Ref("1Gi"),
					Step: convutils.Ref("2Gi"),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("quantityRange.step"))
			Expect(err.Message).To(ContainSubstring("must be less than or equal to quantityRange.max-quantityRange.min"))
		})

		It("accepts quantityRange where step equals max-min", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("0"),
					Max:  convutils.Ref("1Gi"),
					Step: convutils.Ref("1Gi"),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts quantityRange where step has no max to bound it", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("1Gi"),
					Step: convutils.Ref("100Gi"),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("reports min > max first when both min>max and step>span are violated", func() {
			// min=2Gi, max=1Gi, step=100Gi violates both relations; the
			// min<=max error must win, same rationale as range.
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("2Gi"),
					Max:  convutils.Ref("1Gi"),
					Step: convutils.Ref("100Gi"),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("quantityRange.min"))
			Expect(err.Message).To(ContainSubstring("quantityRange.max"))
			Expect(err.Message).NotTo(ContainSubstring("quantityRange.step"))
		})
	})

	Context("range value validation", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)

		It("accepts valid range", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min:  convutils.Ref(float32(11)),
					Max:  convutils.Ref(float32(150)),
					Step: convutils.Ref(float32(4)),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects range where min > max", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min: convutils.Ref(float32(10)),
					Max: convutils.Ref(float32(5)),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must be less than or equal to"))
		})

		It("rejects range with non-positive step", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{Step: convutils.Ref(float32(0))},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must be greater than 0"))
		})

		It("rejects empty range", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("at least one of min/max/step"))
		})

		It("rejects range where step exceeds max-min", func() {
			// min=1, max=10 -> span=9; step=20 leaves only `min` as the single
			// valid value, which is almost certainly a misconfiguration.
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min:  convutils.Ref(float32(1)),
					Max:  convutils.Ref(float32(10)),
					Step: convutils.Ref(float32(20)),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("range.step"))
			Expect(err.Message).To(ContainSubstring("must be less than or equal to range.max-range.min"))
		})

		It("accepts range where step equals max-min", func() {
			// Edge of the allowed region: step == max-min permits exactly two
			// values (min and max), which is still a legitimate policy.
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min:  convutils.Ref(float32(0)),
					Max:  convutils.Ref(float32(10)),
					Step: convutils.Ref(float32(10)),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts range where step has no max to bound it", func() {
			// Without max the span is unbounded above, so step cannot violate
			// the max-min relation.
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min:  convutils.Ref(float32(1)),
					Step: convutils.Ref(float32(100)),
				},
			})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("reports min > max first when both min>max and step>span are violated", func() {
			// min=20, max=10, step=100 violates both min<=max and step<=span.
			// The min/max relation is the more fundamental property, so it
			// must be reported first — a negative span otherwise would make
			// the step error confusing ("step must be less than a negative").
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{
					Min:  convutils.Ref(float32(20)),
					Max:  convutils.Ref(float32(10)),
					Step: convutils.Ref(float32(100)),
				},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("range.min"))
			Expect(err.Message).To(ContainSubstring("range.max"))
			Expect(err.Message).NotTo(ContainSubstring("range.step"))
		})
	})

	Context("match regex validation", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("accepts valid regex", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Match: convutils.Ref("^[a-z]+$")})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects invalid regex", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Match: convutils.Ref("[invalid")})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not a valid regular expression"))
		})
	})

	Context("no-getExpr formatter rejects constraint enforcements", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, noGetFormatter, nil)

		It("rejects required without getExpr or KeyValue", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("require a builtin formatter or a formatter with getExpr"))
		})

		It("accepts set when formatter has setExpr (getExpr is not required for the set value path)", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": "ok"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	Context("set and default require KeyValue or setExpr on the formatter", func() {
		getOnlyFormatter := api.BreakdownFormatter{
			GetExpr: &customGetExpr,
		}
		nonKeyValueBuiltinName := "NotKeyValue"
		builtinNotKeyValue := api.BreakdownFormatter{
			Builtin: &nonKeyValueBuiltinName,
			GetExpr: &customGetExpr,
		}

		It("rejects set when formatter has getExpr but no setExpr and is not KeyValue", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, getOnlyFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "x"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("set is only valid when"))
			Expect(err.Message).To(ContainSubstring("setExpr"))
		})

		It("rejects default when formatter has getExpr but no setExpr and is not KeyValue", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, getOnlyFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "x"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("default is only valid when"))
			Expect(err.Message).To(ContainSubstring("setExpr"))
		})

		It("rejects rule when builtin is unknown (was previously masked by set-enforcement message)", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinNotKeyValue, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"any": "x"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("unknown builtin"))
		})
	})

	// (a) specSelector not in breakdown.SpecSelectors
	Context("specSelector not in breakdown", func() {
		It("rejects when specSelector exists in variation but not in breakdown.SpecSelectors", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1", "orphanSel"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {Query: specSelQuery, ArgsInfo: map[string]api.MappingFieldType{"containerName": api.MAPPINGFIELDTYPE_STRING}},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "orphanSel",
				SelectorParams: map[string]interface{}{},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("specSelector"))
			Expect(err.Message).To(ContainSubstring("not found in the mapping"))
			Expect(err.Details).To(ContainSubstring("sel1"))
		})
	})

	// (b) field with no type defaults to string, rejecting range
	Context("field with empty type defaults to string", func() {
		It("rejects range on field with no type (treated as string)", func() {
			setOnly := api.BreakdownFormatter{SetExpr: &customSetExpr}
			bd := mkBreakdown("fld", "", setOnly, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))},
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not applicable to field type"))
		})
	})

	// (c) enum validation for selector params
	Context("selector param enum validation", func() {
		It("rejects value not in enum", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {
						Query:    `.spec | select(.effect == $ARGS.named.effect)`,
						ArgsInfo: map[string]api.MappingFieldType{"effect": api.MAPPINGFIELDTYPE_STRING},
						ArgsEnum: map[string][]string{"effect": {"NoSchedule", "PreferNoSchedule", "NoExecute"}},
					},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"effect": "InvalidValue"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not one of the allowed values"))
		})

		It("accepts value in enum", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {
						Query:    `.spec | select(.effect == $ARGS.named.effect)`,
						ArgsInfo: map[string]api.MappingFieldType{"effect": api.MAPPINGFIELDTYPE_STRING},
						ArgsEnum: map[string][]string{"effect": {"NoSchedule", "PreferNoSchedule", "NoExecute"}},
					},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"effect": "NoSchedule"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// (d) selector param type validation
	Context("selector param type validation", func() {
		It("rejects non-integer value for integer param", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {
						Query:    `.spec | select(.port == $ARGS.named.port)`,
						ArgsInfo: map[string]api.MappingFieldType{"port": api.MAPPINGFIELDTYPE_INTEGER},
					},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"port": "notAnInt"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected integer"))
		})
	})

	// (e) default enforcement
	Context("default enforcement", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("accepts default with valid value key", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "hello"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects default with wrong key", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Default: map[string]interface{}{"bad": "hello"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must use key"))
		})

		It("rejects default with wrong value type", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "notAnInt"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected integer"))
		})
	})

	// (f) formatter enum validation for set/default
	Context("formatter enum validation", func() {
		It("rejects set value not in formatter enum", func() {
			enumSetExpr := `{"x": $ARGS.named.protocol}`
			enumGetExpr := `.x`
			enumFormatter := api.BreakdownFormatter{
				SetExpr:  &enumSetExpr,
				GetExpr:  &enumGetExpr,
				ArgsInfo: map[string]api.MappingFieldType{"protocol": api.MAPPINGFIELDTYPE_STRING},
				ArgsEnum: map[string][]string{"protocol": {"TCP", "UDP", "SCTP"}},
			}
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, enumFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"protocol": "INVALID"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("not one of the allowed values"))
		})

		It("accepts set value in formatter enum", func() {
			enumSetExpr := `{"x": $ARGS.named.protocol}`
			enumGetExpr := `.x`
			enumFormatter := api.BreakdownFormatter{
				SetExpr:  &enumSetExpr,
				GetExpr:  &enumGetExpr,
				ArgsInfo: map[string]api.MappingFieldType{"protocol": api.MAPPINGFIELDTYPE_STRING},
				ArgsEnum: map[string][]string{"protocol": {"TCP", "UDP", "SCTP"}},
			}
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, enumFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"protocol": "TCP"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// (g) type validation table
	Context("value type validation", func() {
		type typeValEntry struct {
			fieldType   api.MappingFieldType
			value       interface{}
			shouldError bool
			errContains string
		}

		DescribeTable("validates value against type",
			func(e typeValEntry) {
				bd := mkBreakdown("fld", e.fieldType, builtinFormatter, nil)
				rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: e.value}})
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				if e.shouldError {
					Expect(err).NotTo(BeNil())
					Expect(err.Message).To(ContainSubstring(e.errContains))
				} else {
					Expect(err).To(BeNil())
				}
			},
			// string
			Entry("string accepts any string", typeValEntry{api.MAPPINGFIELDTYPE_STRING, "hello", false, ""}),
			Entry("string accepts number as value", typeValEntry{api.MAPPINGFIELDTYPE_STRING, 123, false, ""}),

			// boolean
			Entry("boolean accepts true", typeValEntry{api.MAPPINGFIELDTYPE_BOOLEAN, true, false, ""}),
			Entry("boolean accepts false", typeValEntry{api.MAPPINGFIELDTYPE_BOOLEAN, false, false, ""}),
			Entry("boolean rejects string", typeValEntry{api.MAPPINGFIELDTYPE_BOOLEAN, "yes", true, "expected boolean"}),
			Entry("boolean rejects number", typeValEntry{api.MAPPINGFIELDTYPE_BOOLEAN, 1, true, "expected boolean"}),

			// integer
			Entry("integer accepts int", typeValEntry{api.MAPPINGFIELDTYPE_INTEGER, 42, false, ""}),
			Entry("integer accepts int string", typeValEntry{api.MAPPINGFIELDTYPE_INTEGER, "42", false, ""}),
			Entry("integer rejects float", typeValEntry{api.MAPPINGFIELDTYPE_INTEGER, "3.14", true, "expected integer"}),
			Entry("integer rejects text", typeValEntry{api.MAPPINGFIELDTYPE_INTEGER, "abc", true, "expected integer"}),

			// number
			Entry("number accepts int", typeValEntry{api.MAPPINGFIELDTYPE_NUMBER, 42, false, ""}),
			Entry("number accepts float", typeValEntry{api.MAPPINGFIELDTYPE_NUMBER, 3.14, false, ""}),
			Entry("number accepts float string", typeValEntry{api.MAPPINGFIELDTYPE_NUMBER, "3.14", false, ""}),
			Entry("number rejects text", typeValEntry{api.MAPPINGFIELDTYPE_NUMBER, "abc", true, "expected number"}),

			// quantity
			Entry("quantity accepts k8s quantity", typeValEntry{api.MAPPINGFIELDTYPE_QUANTITY, "256Mi", false, ""}),
			Entry("quantity accepts plain number", typeValEntry{api.MAPPINGFIELDTYPE_QUANTITY, "100", false, ""}),
			Entry("quantity rejects invalid", typeValEntry{api.MAPPINGFIELDTYPE_QUANTITY, "notQuantity", true, "expected quantity"}),

			// ipaddress
			Entry("ip accepts valid IPv4", typeValEntry{api.MAPPINGFIELDTYPE_IPADDRESS, "192.168.1.1", false, ""}),
			Entry("ip rejects invalid", typeValEntry{api.MAPPINGFIELDTYPE_IPADDRESS, "not.an.ip", true, "expected IP address"}),

			// object (no validation)
			Entry("object accepts anything", typeValEntry{api.MAPPINGFIELDTYPE_OBJECT, "whatever", false, ""}),
		)
	})

	// (h) array value validation
	Context("array value validation", func() {
		It("accepts valid integer array", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: []interface{}{1, 2, 3}}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects invalid element in integer array", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: []interface{}{1, "abc", 3}}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected integer"))
		})

		It("accepts comma-separated string for numbers", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_NUMBERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "1.5, 2.5, 3.5"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects invalid element in comma-separated numbers", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_NUMBERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "1.5, abc, 3.5"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected number"))
		})

		It("rejects invalid token in comma-separated integers", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "1, abc, 3"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected integer"))
		})

		It("accepts string array for strings type", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRINGS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: []interface{}{"a", "b"}}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects non-array non-comma scalar for array type", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: true}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("expected array"))
		})
	})

	Context("per-rule enforcement consistency (cross-enforcement)", func() {
		mkRuleMulti := func(field, specSel string, enforcements ...api.PolicyRuleEnforcement) []api.PolicyRule {
			return []api.PolicyRule{{
				Field:          field,
				SpecSelector:   specSel,
				SelectorParams: map[string]interface{}{"containerName": "worker-1"},
				Enforcements:   enforcements,
			}}
		}

		It("accepts set when value matches options and match and length", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			// Every options entry must satisfy peer match/length, not only the set value.
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"hello", "hi"}},
				api.PolicyRuleEnforcement{Match: convutils.Ref("^h.*")},
				api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1)), Max: convutils.Ref(int32(10))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "hello"}},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects set not in options", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"a", "b"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "c"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with options"))
			Expect(err.Message).To(ContainSubstring(`must be one of [a, b]`))
		})

		It("rejects duplicate options enforcements in the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"a", "b"}},
				api.PolicyRuleEnforcement{Options: &[]string{"a", "c"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("duplicate options enforcement"))
		})

		It("rejects rule when an options entry does not satisfy the match enforcement", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"aa", "bb", "cc"}},
				api.PolicyRuleEnforcement{Match: convutils.Ref("^dd$")},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("does not match pattern"))
			Expect(err.Message).To(ContainSubstring(`options value "aa"`))
		})

		It("rejects rule when an options entry violates length bounds", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"ok", "waytoolong"}},
				api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Max: convutils.Ref(int32(2))}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`options value "waytoolong"`))
		})

		It("rejects duplicate length enforcements in the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1)), Max: convutils.Ref(int32(10))}},
				api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Max: convutils.Ref(int32(20))}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("duplicate length enforcement"))
		})

		// A flat (non-itemized) field only has a single scalar target;
		// two sets or a set + default in the same rule race for that
		// target with no way to reconcile them at apply time. Effective
		// rules can still legitimately surface a set and a default for
		// the same field via different scopes (disambiguated by
		// sourceOfRule), but the given rule itself must not carry the
		// conflict.
		It("rejects two set enforcements on a flat field in the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"a", "b", "c"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "a"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "b"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per rule"))
		})

		It("rejects set and default on a flat field in the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Options: &[]string{"x", "y"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "x"}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "y"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per rule"))
		})

		// A setExpr formatter without an ArrayKey is still a flat field:
		// two sets in the same rule race for the same scalar target.
		It("rejects two set enforcements on a flat setExpr field in the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": "a"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": "b"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per rule"))
		})

		It("rejects integer set outside range", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1)), Max: convutils.Ref(float32(10)), Step: convutils.Ref(float32(2))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "4"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
		})

		It("rejects integer set below range min", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(5)), Max: convutils.Ref(float32(10))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "2"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
		})

		It("rejects integer set above range max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1)), Max: convutils.Ref(float32(10))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "11"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
		})

		It("rejects integer set that violates range step from min anchor", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGER, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(0)), Max: convutils.Ref(float32(10)), Step: convutils.Ref(float32(2))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "3"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
		})

		It("accepts integers default when every element satisfies range", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(0)), Max: convutils.Ref(float32(10))}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: []interface{}{1, 2, 3}}},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects integers default when an array element exceeds range max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1)), Max: convutils.Ref(float32(10))}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: []interface{}{1, 2, 11}}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
			Expect(err.Message).To(ContainSubstring("default[2]"))
		})

		It("rejects integers default comma form when a token exceeds range max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1)), Max: convutils.Ref(float32(10))}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "1, 5, 12"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
		})

		It("rejects integers set when an element violates range step", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(0)), Max: convutils.Ref(float32(10)), Step: convutils.Ref(float32(2))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: []interface{}{2, 4, 5}}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
			Expect(err.Message).To(ContainSubstring("set[2]"))
		})

		It("rejects integers default when array length is below elements.min", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_INTEGERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Min: convutils.Ref(int32(4))}},
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(0)), Max: convutils.Ref(float32(100))}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: []interface{}{1, 2, 3}}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("below elements.min"))
		})

		It("rejects numbers default when an element exceeds range max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_NUMBERS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(0)), Max: convutils.Ref(float32(1))}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: []interface{}{0.1, 1.5}}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with range"))
			Expect(err.Message).To(ContainSubstring("default[1]"))
		})

		It("rejects strings array longer than elements.max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRINGS, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Elements: &api.PolicyEnforcementElementsElements{Max: convutils.Ref(int32(2))}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: []interface{}{"a", "b", "c"}}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("exceeds elements.max"))
		})

		It("rejects quantity default below quantityRange.min", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_QUANTITY, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{Min: convutils.Ref("1Gi")}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "100Mi"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with quantityRange"))
			Expect(err.Message).To(ContainSubstring("below minimum"))
		})

		It("rejects quantity set above quantityRange.max", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_QUANTITY, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min: convutils.Ref("100Mi"),
					Max: convutils.Ref("500Mi"),
				}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "1Gi"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with quantityRange"))
			Expect(err.Message).To(ContainSubstring("exceeds maximum"))
		})

		It("rejects quantity default off quantityRange step from zero anchor", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_QUANTITY, builtinFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{QuantityRange: &api.PolicyEnforcementQuantityRangeQuantityRange{
					Min:  convutils.Ref("0"),
					Max:  convutils.Ref("10Gi"),
					Step: convutils.Ref("1Gi"),
				}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "512Mi"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("inconsistent with quantityRange"))
			Expect(err.Message).To(ContainSubstring("multiple"))
		})

		It("rejects custom formatter set parameter not matching match constraint", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Match: convutils.Ref("^z+$")},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"val": "abc"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("does not match pattern"))
		})
	})

	// A formatter with an ArrayKey targets a keyed item inside an array
	// (env vars keyed by name, ports keyed by containerPort, etc.). Multiple
	// set/default enforcements in the same rule are legitimate as long as
	// they address different key values; two enforcements aimed at the same
	// key value are the itemized analog of two sets on a flat scalar and are
	// rejected on the given rule. Effective rules may still surface the same
	// arrayKey twice via different scopes — that layering is disambiguated
	// at effective time via sourceOfRule.
	Context("per-rule set/default uniqueness for itemized fields", func() {
		envSetExpr := `{"env": [{"name": $ARGS.named.name, "value": $ARGS.named.value}]}`
		envGetExpr := `.env[]`
		envFormatter := api.BreakdownFormatter{
			SetExpr:  &envSetExpr,
			GetExpr:  &envGetExpr,
			ArrayKey: &api.FormatterArrayKey{ArgName: "name"},
			ArgsInfo: map[string]api.MappingFieldType{
				"name":  api.MAPPINGFIELDTYPE_STRING,
				"value": api.MAPPINGFIELDTYPE_STRING,
			},
		}

		mkRuleMulti := func(field, specSel string, enforcements ...api.PolicyRuleEnforcement) []api.PolicyRule {
			return []api.PolicyRule{{
				Field:          field,
				SpecSelector:   specSel,
				SelectorParams: map[string]interface{}{"containerName": "worker-1"},
				Enforcements:   enforcements,
			}}
		}

		It("accepts two set enforcements with different arrayKey values", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, envFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY1", "value": "a"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY2", "value": "b"}},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts set and default with different arrayKey values", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, envFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY1", "value": "a"}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{"name": "KEY2", "value": "b"}},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects two set enforcements with the same arrayKey value", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, envFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY1", "value": "a"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY1", "value": "b"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
			Expect(err.Message).To(ContainSubstring(`name="KEY1"`))
		})

		It("rejects set and default with the same arrayKey value", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, envFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"name": "KEY1", "value": "a"}},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{"name": "KEY1", "value": "b"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
			Expect(err.Message).To(ContainSubstring(`name="KEY1"`))
		})

		// Enforcements that omit the arrayKey argument are grouped under
		// the empty key so a rule with two "keyless" enforcements is still
		// caught.
		It("rejects two set enforcements that both omit the arrayKey value", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, envFormatter, nil)
			rules := mkRuleMulti("fld", "sel1",
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"value": "a"}},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{"value": "b"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
		})

		// The KeyValueMap builtin (used by fields like annotations and
		// labels) is itemized in the same sense as an ArrayKey formatter:
		// each set/default targets a specific entry in a map, keyed by
		// the "key" argument. Enforcements with different "key" values
		// address different map entries and must coexist; enforcements
		// with the same "key" would race for the same entry.
		Context("KeyValueMap builtin treats \"key\" as the arrayKey", func() {
			kvm := "KeyValueMap"
			kvmFormatter := api.BreakdownFormatter{Builtin: &kvm}

			It("accepts two set enforcements with different \"key\" values (e.g. two annotations)", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "ccc", DefaultArgName: "ddd"}},
				)
				Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
			})

			It("accepts a set and a default targeting different \"key\" values", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Default: map[string]interface{}{"key": "ccc", DefaultArgName: "ddd"}},
				)
				Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
			})

			It("rejects two set enforcements with the same \"key\" value", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "ccc"}},
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
				Expect(err.Message).To(ContainSubstring(`key="aaa"`))
			})

			It("rejects a set and a default with the same \"key\" value", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Default: map[string]interface{}{"key": "aaa", DefaultArgName: "ccc"}},
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
				Expect(err.Message).To(ContainSubstring(`key="aaa"`))
			})

			// The KeyValueMap "no key" fan-out form ({ value: <v> }) applies
			// the same value to every existing map key. Two such enforcements
			// on the same rule set the same target map at once and are
			// contradictory; group them under the empty key so the duplicate
			// is caught.
			It("rejects two set enforcements that both omit \"key\" (fan-out form)", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "aaa"}},
					api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "bbb"}},
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
			})

			// A keyless fan-out ({ value: <v> }) is applied to every existing
			// map key at runtime, so it overlaps with any explicit-key
			// enforcement in the same rule regardless of the order the two
			// appear in. Neither can hide under a separate group.
			It("rejects an explicit \"key\" set followed by a keyless fan-out default", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{"key": "aaa", DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "ccc"}},
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
			})

			It("rejects a keyless fan-out set followed by an explicit \"key\" default", func() {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, kvmFormatter, nil)
				rules := mkRuleMulti("fld", "sel1",
					api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "bbb"}},
					api.PolicyRuleEnforcement{Default: map[string]interface{}{"key": "aaa", DefaultArgName: "ccc"}},
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("only one set or default is allowed per arrayKey value"))
			})
		})
	})

	// the runtime resolver (gvk_compliance.ResolveRules) drops a rule when its
	// spec selector has an empty/missing query. validateRules used to accept
	// that policy at create time and only the apply-time LogBrokenRule warning
	// would expose the drift. these tests pin the explicit rejection.
	Context("spec selector query (closes ResolveRules empty-Query gap)", func() {
		It("rejects rule when spec selector has an empty Query", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {Query: ""},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`specSelector "sel1" has an empty query`))
		})

		It("rejects rule when spec selector Query is whitespace only", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {Query: "   \t  "},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("has an empty query"))
		})
	})

	// gvk_compliance.ResolveRules concatenates specSelector | itemSelector and
	// compiles the combined JQ; validateRules used to only Parse each piece
	// independently and silently swallow the error. these tests pin that the
	// composed-and-compiled form is now validated.
	Context("composed selector JQ (closes ResolveRules compile gap)", func() {
		It("rejects malformed spec selector query", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {Query: "(("},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("composed selector query failed to parse"))
		})

		It("rejects spec selector query that parses but fails to compile (undefined function)", func() {
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {Query: "no_such_function"},
				},
			}
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("composed selector query failed to compile"))
		})

		It("rejects malformed item selector query when piped after spec selector", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, map[string]string{
				"iSelBad": "((",
			})
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSelBad"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("composed selector query failed to parse"))
		})

		It("accepts a valid composed spec | item query", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, map[string]string{
				"iSel1": `(.items // [])[]`,
			})
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1", ItemSelector: convutils.Ref("iSel1"),
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// when a Builtin name is set, the runtime only knows KeyValue and
	// KeyValueMap; everything else becomes a broken rule at apply time. for
	// non-set/default enforcements validateRules used to accept any non-nil
	// Builtin (constraint enforcements only checked "GetExpr or Builtin must
	// be set"). these tests pin the new explicit rejection of unknown names.
	Context("formatter builtin name (closes unknown-builtin gap)", func() {
		It("rejects required enforcement with an unknown builtin formatter", func() {
			unknown := "Foo"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{Builtin: &unknown}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`unknown builtin "Foo"`))
			Expect(err.Details).To(ContainSubstring("KeyValue"))
			Expect(err.Details).To(ContainSubstring("KeyValueMap"))
		})

		It("rejects match enforcement with an unknown builtin formatter", func() {
			unknown := "Bar"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{Builtin: &unknown}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Match: convutils.Ref("^a$")})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("unknown builtin"))
		})

		It("accepts the KeyValue builtin", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts the KeyValueMap builtin (now recognized alongside KeyValue)", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts set against the KeyValueMap builtin (was previously misreported as 'set is only valid when KeyValue or setExpr')", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "hello"}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		// KeyValueMap's compiled set expression is:
		//   { ($ARGS.named.key // "value"): $ARGS.named.value }
		// It therefore accepts either:
		//   * { value: <v> }              — fanned to every existing key (CopyTargetKeys)
		//   * { key: <k>, value: <v> }    — sets a single specific key
		// The validator must accept both shapes; previously it treated KeyValueMap
		// like KeyValue and rejected the { key, value } form with
		// "must contain exactly one entry with key value, got 2 entries".
		It("accepts KeyValueMap set with { key, value } pair", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key":          "app",
				DefaultArgName: "dynamo",
			}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts KeyValueMap set with just { value } (fan-to-all-keys form)", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				DefaultArgName: "shared",
			}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects KeyValueMap set that omits the value key", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key": "app",
			}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`must include the "value" key`))
		})

		It("rejects KeyValueMap set with unexpected extra keys", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key":          "app",
				DefaultArgName: "dynamo",
				"unexpected":   "x",
			}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("unexpected"))
		})

		// JQ object construction requires the key expression to evaluate to a
		// string. The runtime would otherwise produce a confusing error at
		// apply time; validate it at the API layer.
		It("rejects KeyValueMap set when key is not a string", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key":          123,
				DefaultArgName: "dynamo",
			}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring(`"key" value to be a string`))
		})

		It("accepts KeyValueMap set with a string key (positive case for the type check)", func() {
			kvm := "KeyValueMap"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_OBJECT, api.BreakdownFormatter{Builtin: &kvm}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key":          "app",
				DefaultArgName: "dynamo",
			}})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		// KeyValue's validator behavior is unchanged: exactly one "value" entry.
		It("rejects KeyValue set with extra key/value pair", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{
				"key":          "app",
				DefaultArgName: "dynamo",
			}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must contain exactly one entry"))
		})
	})

	// the runtime compiles getExpr (no $ARGS) and setExpr (with $ARGS).
	// validateRules used to skip getExpr entirely and only Parse setExpr for
	// argument extraction, so a malformed expression on either side would only
	// be reported at apply time via LogBrokenRule. these tests pin the new
	// parse + compile checks.
	Context("formatter getExpr / setExpr compile (closes getExpr+setExpr gaps)", func() {
		It("rejects formatter with malformed getExpr (parse failure)", func() {
			badGet := "(("
			set := `{"x": $ARGS.named.val}`
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{
				GetExpr:  &badGet,
				SetExpr:  &set,
				ArgsInfo: map[string]api.MappingFieldType{"val": api.MAPPINGFIELDTYPE_STRING},
			}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("formatter getExpr failed to parse"))
		})

		It("rejects formatter whose getExpr parses but fails to compile (uses $ARGS, which getExpr is compiled without)", func() {
			// getExpr is compiled without WithVariables(["$ARGS"]) at runtime,
			// so an expression that references $ARGS will compile-fail. it
			// would parse cleanly, so this case is exclusive to the new
			// compile step.
			badGet := "$ARGS.named.val"
			set := `{"x": $ARGS.named.val}`
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{
				GetExpr:  &badGet,
				SetExpr:  &set,
				ArgsInfo: map[string]api.MappingFieldType{"val": api.MAPPINGFIELDTYPE_STRING},
			}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("formatter getExpr failed to compile"))
		})

		It("rejects formatter with malformed setExpr (parse failure)", func() {
			get := ".x"
			badSet := "(("
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{
				GetExpr: &get,
				SetExpr: &badSet,
			}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"any": "x"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("formatter setExpr failed to parse"))
		})

		It("rejects formatter whose setExpr parses but fails to compile (undefined function reference)", func() {
			// gojq.Parse accepts unknown function names; gojq.Compile resolves
			// them and fails. validateRules used to only Parse setExpr.
			get := ".x"
			badSet := `{"x": no_such_function(.)}`
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{
				GetExpr: &get,
				SetExpr: &badSet,
			}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"any": "x"}})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("formatter setExpr failed to compile"))
		})

		It("accepts a formatter whose getExpr/setExpr both parse and compile cleanly", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, customFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// karta's ValidateParsedJQ rejects JQ that mutates the input or may produce
	// unbounded output. these tests verify that all three JQ-bearing paths in the
	// policy validator (specSelector query, formatter getExpr, formatter setExpr)
	// surface that rejection at create/update time rather than silently applying
	// a dangerous rule.
	Context("unsafe JQ expressions are rejected (protect from misuse)", func() {
		mkCustomFormatter := func(getExpr, setExpr string) api.BreakdownFormatter {
			return api.BreakdownFormatter{
				GetExpr: &getExpr,
				SetExpr: &setExpr,
			}
		}

		DescribeTable("rejects dangerous JQ in formatter getExpr",
			func(expr string, errFragment string) {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, mkCustomFormatter(expr, `{"x":$ARGS.named.val}`), nil)
				rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("is unsafe"))
				Expect(err.Message).To(ContainSubstring(errFragment))
			},
			Entry("del function", `del(.x)`, "del"),
			Entry("recursive descent ..", `.. | .x`, "recursive descent"),
			Entry("recurse function", `recurse | .x`, "recurse"),
			Entry("paths function", `paths`, "paths"),
			Entry("range function", `range(1000000)`, "range"),
			Entry("walk function", `walk(.)`, "walk"),
			Entry("repeat function", `first(repeat(. + 1))`, "repeat"),
		)

		DescribeTable("rejects dangerous JQ in formatter setExpr",
			func(expr string, errFragment string) {
				bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, mkCustomFormatter(`.x`, expr), nil)
				rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Set: map[string]interface{}{"any": "v"}})
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("is unsafe"))
				Expect(err.Message).To(ContainSubstring(errFragment))
			},
			Entry("assignment operator", `.x = 1`, "modifying operator"),
			Entry("update operator", `.x |= . + 1`, "modifying operator"),
			Entry("del function", `del(.x)`, "del"),
			Entry("recursive descent ..", `.. | .x`, "recursive descent"),
		)

		DescribeTable("rejects dangerous JQ in spec selector query",
			func(query string, errFragment string) {
				bd := &api.BreakdownBreakdown{
					Fields: map[string][]api.BreakdownFieldVariation{
						"fld": {{SpecSelectors: []string{"sel1"}, Type: api.MAPPINGFIELDTYPE_STRING, Formatter: builtinFormatter}},
					},
					SpecSelectors: map[string]api.BreakdownSpecSelector{
						"sel1": {Query: query},
					},
				}
				rules := []api.PolicyRule{{
					Field: "fld", SpecSelector: "sel1",
					Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
				}}
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("is unsafe"))
				Expect(err.Message).To(ContainSubstring(errFragment))
			},
			Entry("del function", `del(.x)`, "del"),
			Entry("recursive descent ..", `.. | .x`, "recursive descent"),
			Entry("recurse function", `recurse | .x`, "recurse"),
			Entry("paths function", `paths`, "paths"),
		)

		It("accepts valid safe JQ in all positions", func() {
			safeQuery := `.spec.containers[] | select(.name == $ARGS.named.containerName)`
			bd := &api.BreakdownBreakdown{
				Fields: map[string][]api.BreakdownFieldVariation{
					"fld": {{
						SpecSelectors: []string{"sel1"},
						Type:          api.MAPPINGFIELDTYPE_STRING,
						Formatter:     mkCustomFormatter(`.x`, `{"x":$ARGS.named.val}`),
					}},
				},
				SpecSelectors: map[string]api.BreakdownSpecSelector{
					"sel1": {
						Query:    safeQuery,
						ArgsInfo: map[string]api.MappingFieldType{"containerName": api.MAPPINGFIELDTYPE_STRING},
						ArgsEnum: map[string][]string{},
					},
				},
			}
			rules := []api.PolicyRule{{
				Field:          "fld",
				SpecSelector:   "sel1",
				SelectorParams: map[string]interface{}{"containerName": "worker"},
				Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}}
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// validateRules walks every rule and every check and appends each
	// finding to the caller-supplied collector. Pure ruleIssues unit
	// tests (the contract of add/addf/addWithDetails/full/toError/all)
	// live in rule_issues_test.go; this Context exercises the
	// integration: that validateRules' loop and per-rule pipeline use
	// the collector correctly across multiple rules and within a single
	// rule.
	Context("validateRules + ruleIssues integration", func() {
		// helper: collect every issue (no cap) and return them as messages
		collectMsgs := func(bd *api.BreakdownBreakdown, rules []api.PolicyRule) []string {
			issues := pa.validateRules(nil, bd, rules, newRuleIssues(0))
			out := make([]string, 0, len(issues.all()))
			for _, e := range issues.all() {
				out = append(out, e.Message)
			}
			return out
		}

		It("collects issues from every rule when multiple rules each have a different problem", func() {
			bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := []api.PolicyRule{
				{
					// rule[0]: unknown field
					Field: "unknownField", SpecSelector: "sel1",
					SelectorParams: map[string]interface{}{"containerName": "w"},
					Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
				},
				{
					// rule[1]: unknown specSelector
					Field: "fieldA", SpecSelector: "badSel",
					SelectorParams: map[string]interface{}{"containerName": "w"},
					Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
				},
				{
					// rule[2]: structurally fine — should not generate an issue
					Field: "fieldA", SpecSelector: "sel1",
					SelectorParams: map[string]interface{}{"containerName": "w"},
					Enforcements:   []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
				},
			}
			msgs := collectMsgs(bd, rules)
			Expect(msgs).To(HaveLen(2))
			Expect(msgs[0]).To(ContainSubstring("rule[0]"))
			Expect(msgs[0]).To(ContainSubstring(`field "unknownField" not found`))
			Expect(msgs[1]).To(ContainSubstring("rule[1]"))
			Expect(msgs[1]).To(ContainSubstring("specSelector"))
		})

		It("collects independent issues within a single rule (formatter + enforcement)", func() {
			// Builtin "Foo" is not in builtinFormatters → formatter issue.
			// Range enforcement on a STRING field → enforcement-vs-type issue.
			unknown := "Foo"
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, api.BreakdownFormatter{Builtin: &unknown}, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))},
			})
			msgs := collectMsgs(bd, rules)
			Expect(msgs).To(HaveLen(2))
			Expect(strings.Join(msgs, "\n")).To(ContainSubstring(`unknown builtin "Foo"`))
			Expect(strings.Join(msgs, "\n")).To(ContainSubstring("not applicable to field type"))
		})

		It("collects issues from multiple enforcements within the same rule", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := []api.PolicyRule{{
				Field: "fld", SpecSelector: "sel1",
				SelectorParams: map[string]interface{}{"containerName": "w"},
				Enforcements: []api.PolicyRuleEnforcement{
					// enforcement[0]: empty (no field set) → "exactly one non-nil field"
					{},
					// enforcement[1]: range on a STRING field → "not applicable"
					{Range: &api.PolicyEnforcementRangeRange{Min: convutils.Ref(float32(1))}},
					// enforcement[2]: structurally fine
					{Required: convutils.Ref(true)},
				},
			}}
			joined := strings.Join(collectMsgs(bd, rules), "\n")
			Expect(joined).To(ContainSubstring("enforcement[0]"))
			Expect(joined).To(ContainSubstring("exactly one non-nil field"))
			Expect(joined).To(ContainSubstring("enforcement[1]"))
			Expect(joined).To(ContainSubstring("not applicable to field type"))
			Expect(joined).NotTo(ContainSubstring("enforcement[2]"))
		})

		It("keeps validating subsequent rules after a duplicate is reported", func() {
			// rule[0] is fine, rule[1] is a duplicate, rule[2] has an
			// independent unknown-field issue. Both findings must be in
			// the collector after the call.
			bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			ok := api.PolicyRuleEnforcement{Required: convutils.Ref(true)}
			rules := []api.PolicyRule{
				{Field: "fieldA", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "fieldA", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
			}
			joined := strings.Join(collectMsgs(bd, rules), "\n")
			Expect(joined).To(ContainSubstring("duplicate rule"))
			Expect(joined).To(ContainSubstring(`field "missing" not found`))
		})

		It("returns nil when every rule passes", func() {
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			rules := []api.PolicyRule{
				mkRule("fld", "sel1", api.PolicyRuleEnforcement{Required: convutils.Ref(true)})[0],
				// note: distinct selectorParams so it isn't a duplicate of the first
				{Field: "fld", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "other"}, Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})
	})

	// maxErrors caps how many findings the collector retains. Once the cap
	// is hit, every subsequent add is dropped and full() returns true so
	// validateRules and the breakdown generator can stop. create/update
	// pass maxErrors=1 to recover the legacy "fail on first issue" behavior.
	Context("maxErrors cap", func() {
		It("stops collecting once maxErrors is reached", func() {
			bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			ok := api.PolicyRuleEnforcement{Required: convutils.Ref(true)}
			rules := []api.PolicyRule{
				{Field: "missing1", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing2", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing3", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
			}
			issues := pa.validateRules(nil, bd, rules, newRuleIssues(2))
			Expect(issues.all()).To(HaveLen(2))
			Expect(issues.full()).To(BeTrue())
			// only the first two rule-level issues are recorded — the
			// third rule was never visited because the loop checks full()
			// before calling validateOneRule.
			Expect(issues.all()[0].Message).To(ContainSubstring(`field "missing1" not found`))
			Expect(issues.all()[1].Message).To(ContainSubstring(`field "missing2" not found`))
		})

		It("with maxErrors=1 returns immediately on the first finding (legacy create/update behavior)", func() {
			bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			ok := api.PolicyRuleEnforcement{Required: convutils.Ref(true)}
			rules := []api.PolicyRule{
				{Field: "missing1", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing2", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
			}
			issues := pa.validateRules(nil, bd, rules, newRuleIssues(1))
			Expect(issues.all()).To(HaveLen(1))
			Expect(issues.full()).To(BeTrue())
			Expect(issues.toError().Message).To(ContainSubstring(`field "missing1" not found`))
		})

		It("with maxErrors=0 collects every finding (unlimited)", func() {
			bd := mkBreakdown("fieldA", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)
			ok := api.PolicyRuleEnforcement{Required: convutils.Ref(true)}
			rules := []api.PolicyRule{
				{Field: "missing1", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing2", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
				{Field: "missing3", SpecSelector: "sel1", SelectorParams: map[string]interface{}{"containerName": "w"}, Enforcements: []api.PolicyRuleEnforcement{ok}},
			}
			issues := pa.validateRules(nil, bd, rules, newRuleIssues(0))
			Expect(issues.all()).To(HaveLen(3))
			Expect(issues.full()).To(BeFalse())
		})
	})

	// blocked is the inverse of required: it rejects any value present in the field.
	// Pin the structural rules around it so the validator can't quietly relax them.
	Context("blocked enforcement", func() {
		bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, builtinFormatter, nil)

		It("accepts blocked: true alone", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts blocked: false alone", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Blocked: convutils.Ref(false)})
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects an enforcement that has both blocked and required set (single-enforcement rule)", func() {
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{
				Blocked:  convutils.Ref(true),
				Required: convutils.Ref(true),
			})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("exactly one non-nil field"))
		})

		It("rejects blocked: true + required: true across two enforcements on the same rule", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
				api.PolicyRuleEnforcement{Required: convutils.Ref(true)},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("blocked is incompatible with required"))
		})

		It("rejects blocked: true + set across two enforcements on the same rule", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
				api.PolicyRuleEnforcement{Set: map[string]interface{}{DefaultArgName: "x"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("blocked is incompatible with set"))
		})

		It("rejects blocked: true + default across two enforcements on the same rule", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
				api.PolicyRuleEnforcement{Default: map[string]interface{}{DefaultArgName: "x"}},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("blocked is incompatible with default"))
		})

		It("accepts blocked: false alongside required: true", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(false)},
				api.PolicyRuleEnforcement{Required: convutils.Ref(true)},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("accepts blocked: true alongside required: false", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
				api.PolicyRuleEnforcement{Required: convutils.Ref(false)},
			)
			Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).To(BeNil())
		})

		It("rejects duplicate blocked enforcements in the same rule", func() {
			rules := mkRuleWithEnforcements("fld", "sel1",
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
				api.PolicyRuleEnforcement{Blocked: convutils.Ref(false)},
			)
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("duplicate blocked enforcement"))
		})

		// blocked=true means "the field must be empty", so every peer
		// enforcement in the same rule is either contradictory or vacuous;
		// exercise a representative subset of the enforcement kinds to
		// pin down that behavior.
		DescribeTable("rejects blocked: true paired with any other enforcement type",
			func(peer api.PolicyRuleEnforcement, peerKind string) {
				rules := mkRuleWithEnforcements("fld", "sel1",
					api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)},
					peer,
				)
				err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("blocked is incompatible with " + peerKind))
			},
			Entry("match", api.PolicyRuleEnforcement{Match: convutils.Ref("^abc$")}, EnforcementMatch),
			Entry("options", api.PolicyRuleEnforcement{Options: &[]string{"a", "b"}}, EnforcementOptions),
			Entry("length", api.PolicyRuleEnforcement{Length: &api.PolicyEnforcementLengthLength{Min: convutils.Ref(int32(1))}}, EnforcementLength),
		)

		It("accepts blocked on every field type (allTypes)", func() {
			for _, ft := range []api.MappingFieldType{
				api.MAPPINGFIELDTYPE_STRING,
				api.MAPPINGFIELDTYPE_INTEGER,
				api.MAPPINGFIELDTYPE_NUMBER,
				api.MAPPINGFIELDTYPE_BOOLEAN,
				api.MAPPINGFIELDTYPE_QUANTITY,
				api.MAPPINGFIELDTYPE_OBJECT,
				api.MAPPINGFIELDTYPE_STRINGS,
				api.MAPPINGFIELDTYPE_INTEGERS,
				api.MAPPINGFIELDTYPE_NUMBERS,
				api.MAPPINGFIELDTYPE_NAMES_ARRAY,
			} {
				bd := mkBreakdown("fld", ft, builtinFormatter, nil)
				rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)})
				Expect(pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()).
					To(BeNil(), "blocked unexpectedly rejected on field type %q", ft)
			}
		})

		It("rejects blocked when the formatter has neither getExpr nor a builtin", func() {
			// blocked needs the formatter's getExpr to evaluate presence at
			// runtime — the same precondition required uses. The shared
			// "constraint enforcements require a builtin formatter or a formatter
			// with getExpr" check covers it.
			bd := mkBreakdown("fld", api.MAPPINGFIELDTYPE_STRING, noGetFormatter, nil)
			rules := mkRule("fld", "sel1", api.PolicyRuleEnforcement{Blocked: convutils.Ref(true)})
			err := pa.validateRules(nil, bd, rules, newRuleIssues(0)).toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("require a builtin formatter or a formatter with getExpr"))
		})
	})
})

var _ = Describe("possibleEnforcementsForField", func() {
	builtin := api.BreakdownFormatter{Builtin: convutils.Ref("KeyValueMap")}

	// Reference set of the enforcements always allowed regardless of type
	// (mirrors enforcementsTable's allTypes entries — set/default/required/
	// blocked). Kept as a slice so callers can spread it before the type-
	// specific extras and have the expected slice match the order
	// possibleEnforcementsForField produces.
	always := []api.MappingPossibleEnforcement{
		api.MAPPINGPOSSIBLEENFORCEMENT_SET,
		api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
		api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
		api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
	}

	// The cases below intentionally encode the (fieldType -> extras) map
	// independently of enforcementsTable so that any future drift between
	// the table and the helper is caught here instead of silently agreeing
	// with itself.
	DescribeTable("derives the per-type extras (no itemSelectors)",
		func(fieldType api.MappingFieldType, extras []api.MappingPossibleEnforcement) {
			got := possibleEnforcementsForField(fieldType, false, builtin)
			Expect(got).To(Equal(append(append([]api.MappingPossibleEnforcement{}, always...), extras...)))
		},
		Entry("string", api.MAPPINGFIELDTYPE_STRING, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
		}),
		Entry("boolean", api.MAPPINGFIELDTYPE_BOOLEAN, []api.MappingPossibleEnforcement{}),
		Entry("integer", api.MAPPINGFIELDTYPE_INTEGER, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_RANGE,
		}),
		Entry("number", api.MAPPINGFIELDTYPE_NUMBER, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_RANGE,
		}),
		Entry("quantity", api.MAPPINGFIELDTYPE_QUANTITY, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_QUANTITY_RANGE,
		}),
		Entry("ipaddress", api.MAPPINGFIELDTYPE_IPADDRESS, []api.MappingPossibleEnforcement{}),
		Entry("integers", api.MAPPINGFIELDTYPE_INTEGERS, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_RANGE,
			api.MAPPINGPOSSIBLEENFORCEMENT_ELEMENTS,
		}),
		Entry("numbers", api.MAPPINGFIELDTYPE_NUMBERS, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_RANGE,
			api.MAPPINGPOSSIBLEENFORCEMENT_ELEMENTS,
		}),
		Entry("strings", api.MAPPINGFIELDTYPE_STRINGS, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
			api.MAPPINGPOSSIBLEENFORCEMENT_ELEMENTS,
		}),
		Entry("namesArray", api.MAPPINGFIELDTYPE_NAMES_ARRAY, []api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
		}),
		Entry("object", api.MAPPINGFIELDTYPE_OBJECT, []api.MappingPossibleEnforcement{}),
	)

	It("appends items only when the field declares itemSelectors", func() {
		// items must be present regardless of fieldType when itemSelectors
		// exist (matches the special-case at validator line ~172).
		got := possibleEnforcementsForField(api.MAPPINGFIELDTYPE_BOOLEAN, true, builtin)
		Expect(got).To(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS))

		got = possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, builtin)
		Expect(got).To(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS))

		got = possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, false, builtin)
		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS))
	})

	It("treats an empty fieldType the same as string (defensive fallback)", func() {
		// matches effectiveFieldType — same convention the validator now
		// uses, and the same default used elsewhere for missing arg types
		// (selector args, formatter args).
		Expect(possibleEnforcementsForField("", false, builtin)).
			To(Equal(possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, false, builtin)))
	})

	It("emits only enforcements that are in the AllowedMappingPossibleEnforcement enum", func() {
		// guards against the helper producing kinds outside the published
		// enum if the local string constants and the OpenAPI enum drift.
		allowed := map[api.MappingPossibleEnforcement]struct{}{}
		for _, v := range api.AllowedMappingPossibleEnforcementEnumValues {
			allowed[v] = struct{}{}
		}
		for _, ft := range api.AllowedMappingFieldTypeEnumValues {
			for _, kind := range possibleEnforcementsForField(ft, true, builtin) {
				_, ok := allowed[kind]
				Expect(ok).To(BeTrue(), "enforcement %q for type %q not in enum", kind, ft)
			}
		}
	})

	It("omits set/default for a get-only formatter (runai-metadata companion)", func() {
		// annotations-value / labels-value: GetExpr ".", no SetExpr, not a
		// builtin. validateEnforcementValues rejects set/default here, so
		// field-info must not advertise them.
		getOnly := api.BreakdownFormatter{GetExpr: convutils.Ref(".")}

		got := possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, getOnly)

		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_SET))
		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT))
		// read-based enforcements survive because there is a getExpr.
		Expect(got).To(ContainElements(
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
			api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS,
		))
	})

	It("keeps set/default for a builtin formatter", func() {
		got := possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, builtin)

		Expect(got).To(ContainElements(
			api.MAPPINGPOSSIBLEENFORCEMENT_SET,
			api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
			api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS,
		))
	})

	It("keeps set/default for a setExpr formatter", func() {
		setOnly := api.BreakdownFormatter{SetExpr: convutils.Ref(`{"x":$ARGS.named.value}`)}

		got := possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, setOnly)

		Expect(got).To(ContainElements(
			api.MAPPINGPOSSIBLEENFORCEMENT_SET,
			api.MAPPINGPOSSIBLEENFORCEMENT_DEFAULT,
		))
		// no getExpr and not a builtin → the read-based enforcements are dropped.
		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED))
		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_MATCH))
		Expect(got).NotTo(ContainElement(api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS))
	})

	It("drops everything for a formatter that can neither read nor write", func() {
		empty := api.BreakdownFormatter{}

		Expect(possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, empty)).To(BeEmpty())
	})

	It("preserves the input order of the surviving enforcements", func() {
		getOnly := api.BreakdownFormatter{GetExpr: convutils.Ref(".")}

		got := possibleEnforcementsForField(api.MAPPINGFIELDTYPE_STRING, true, getOnly)

		Expect(got).To(Equal([]api.MappingPossibleEnforcement{
			api.MAPPINGPOSSIBLEENFORCEMENT_REQUIRED,
			api.MAPPINGPOSSIBLEENFORCEMENT_BLOCKED,
			api.MAPPINGPOSSIBLEENFORCEMENT_MATCH,
			api.MAPPINGPOSSIBLEENFORCEMENT_OPTIONS,
			api.MAPPINGPOSSIBLEENFORCEMENT_LENGTH,
			api.MAPPINGPOSSIBLEENFORCEMENT_ITEMS,
		}))
	})
})
