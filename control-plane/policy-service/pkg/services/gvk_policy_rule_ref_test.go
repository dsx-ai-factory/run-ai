package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("PolicyRuleSet", func() {

	Describe("mergeGvkPolicyRulesByRef", func() {
		It("replaces matching rules and appends new ones", func() {
			existing := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "b", SpecSelector: "s2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
			}
			incoming := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}}},
				{Field: "c", SpecSelector: "s3", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}

			out := mergeGvkPolicyRulesByRef(existing, incoming)

			Expect(out).To(HaveLen(3))
			Expect(*out[0].Enforcements[0].Required).To(BeFalse())
			Expect(*out[1].Enforcements[0].Required).To(BeFalse())
			Expect(out[2].Field).To(Equal("c"))
		})

		It("preserves order when replacing in-place", func() {
			existing := []api.PolicyRule{
				{Field: "x", SpecSelector: "s1"},
				{Field: "y", SpecSelector: "s2"},
				{Field: "z", SpecSelector: "s3"},
			}
			incoming := []api.PolicyRule{
				{Field: "y", SpecSelector: "s2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}

			out := mergeGvkPolicyRulesByRef(existing, incoming)

			Expect(out).To(HaveLen(3))
			Expect(out[0].Field).To(Equal("x"))
			Expect(out[1].Field).To(Equal("y"))
			Expect(out[2].Field).To(Equal("z"))
			Expect(out[1].Enforcements).To(HaveLen(1))
		})
	})

	Describe("removeGvkPolicyRulesByRefs", func() {
		It("removes matching rules", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "b", SpecSelector: "s2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			rm := []api.PolicyRuleRef{{Field: "a", SpecSelector: "s1"}}

			out, err := removeGvkPolicyRulesByRefs(rules, rm)

			Expect(err).To(BeNil())
			Expect(out).To(HaveLen(1))
			Expect(out[0].Field).To(Equal("b"))
		})

		It("preserves order of remaining rules", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
				{Field: "c", SpecSelector: "s3"},
				{Field: "d", SpecSelector: "s4"},
			}
			rm := []api.PolicyRuleRef{{Field: "b", SpecSelector: "s2"}}

			out, err := removeGvkPolicyRulesByRefs(rules, rm)

			Expect(err).To(BeNil())
			Expect(out).To(HaveLen(3))
			Expect(out[0].Field).To(Equal("a"))
			Expect(out[1].Field).To(Equal("c"))
			Expect(out[2].Field).To(Equal("d"))
		})

		It("returns error when rule ref is not found", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
			}
			rm := []api.PolicyRuleRef{{Field: "missing", SpecSelector: "x"}}

			_, err := removeGvkPolicyRulesByRefs(rules, rm)

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(400)))
		})

		It("returns error when removing a mix of existing and non-existing refs", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
			}
			rm := []api.PolicyRuleRef{
				{Field: "a", SpecSelector: "s1"},
				{Field: "nonexistent", SpecSelector: "s9"},
			}

			_, err := removeGvkPolicyRulesByRefs(rules, rm)

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(400)))
		})

		It("returns all rules removed when every rule is deleted", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
			}
			rm := []api.PolicyRuleRef{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
			}

			out, err := removeGvkPolicyRulesByRefs(rules, rm)

			Expect(err).To(BeNil())
			Expect(out).To(BeEmpty())
		})
	})

	Describe("ref key variants", func() {
		It("matches by field + specSelector only", func() {
			rules := []api.PolicyRule{
				{Field: "f1", SpecSelector: "sel1", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
				{Field: "f2", SpecSelector: "sel2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}},
			}
			set := NewPolicyRuleSet(rules)

			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1"})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel2"})).To(BeFalse())

			Expect(set.Delete(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1"})).To(BeTrue())
			Expect(set.Rules()).To(HaveLen(1))
			Expect(set.Rules()[0].Field).To(Equal("f2"))
		})

		It("matches by field + specSelector + itemSelector", func() {
			rules := []api.PolicyRule{
				{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1")},
				{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item2")},
				{Field: "f1", SpecSelector: "sel1"},
			}
			set := NewPolicyRuleSet(rules)

			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1")})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item2")})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1"})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item3")})).To(BeFalse())

			Expect(set.Delete(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1")})).To(BeTrue())
			Expect(set.Rules()).To(HaveLen(2))
			Expect(set.Rules()[0].ItemSelector).To(Equal(convutils.Ref("item2")))
		})

		It("matches by field + specSelector + itemSelector + selectorParams", func() {
			params1 := map[string]interface{}{"env": "prod", "region": "us"}
			params2 := map[string]interface{}{"env": "staging"}

			rules := []api.PolicyRule{
				{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: params1},
				{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: params2},
				{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1")},
			}
			set := NewPolicyRuleSet(rules)

			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: params1})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: params2})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1")})).To(BeTrue())

			paramsReordered := map[string]interface{}{"region": "us", "env": "prod"}
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: paramsReordered})).To(BeTrue())

			paramsDifferent := map[string]interface{}{"env": "dev"}
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: paramsDifferent})).To(BeFalse())

			Expect(set.Delete(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", ItemSelector: convutils.Ref("item1"), SelectorParams: params1})).To(BeTrue())
			Expect(set.Rules()).To(HaveLen(2))
		})

		It("matches by field + specSelector + selectorParams (no itemSelector)", func() {
			params := map[string]interface{}{"key": "value"}

			rules := []api.PolicyRule{
				{Field: "f1", SpecSelector: "sel1", SelectorParams: params},
				{Field: "f1", SpecSelector: "sel1"},
			}
			set := NewPolicyRuleSet(rules)

			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", SelectorParams: params})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1"})).To(BeTrue())

			Expect(set.Delete(api.PolicyRuleRef{Field: "f1", SpecSelector: "sel1", SelectorParams: params})).To(BeTrue())
			out := set.Rules()
			Expect(out).To(HaveLen(1))
			Expect(out[0].SelectorParams).To(BeNil())
		})
	})

	Describe("PolicyRuleSet Set/Delete/Has", func() {
		It("supports add, lookup, delete, and not-found", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
			}
			set := NewPolicyRuleSet(rules)

			Expect(set.Has(api.PolicyRuleRef{Field: "a", SpecSelector: "s1"})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "b", SpecSelector: "s2"})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "c", SpecSelector: "s3"})).To(BeFalse())

			set.Set(api.PolicyRule{Field: "c", SpecSelector: "s3", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}})
			Expect(set.Has(api.PolicyRuleRef{Field: "c", SpecSelector: "s3"})).To(BeTrue())

			Expect(set.Delete(api.PolicyRuleRef{Field: "b", SpecSelector: "s2"})).To(BeTrue())
			Expect(set.Has(api.PolicyRuleRef{Field: "b", SpecSelector: "s2"})).To(BeFalse())

			Expect(set.Delete(api.PolicyRuleRef{Field: "missing", SpecSelector: "x"})).To(BeFalse())

			out := set.Rules()
			Expect(out).To(HaveLen(2))
			Expect(out[0].Field).To(Equal("a"))
			Expect(out[1].Field).To(Equal("c"))
		})

		It("replaces a rule in-place preserving position", func() {
			rules := []api.PolicyRule{
				{Field: "a", SpecSelector: "s1"},
				{Field: "b", SpecSelector: "s2"},
				{Field: "c", SpecSelector: "s3"},
			}
			set := NewPolicyRuleSet(rules)

			set.Set(api.PolicyRule{Field: "b", SpecSelector: "s2", Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}}})

			out := set.Rules()
			Expect(out).To(HaveLen(3))
			Expect(out[0].Field).To(Equal("a"))
			Expect(out[1].Field).To(Equal("b"))
			Expect(out[2].Field).To(Equal("c"))
			Expect(out[1].Enforcements).To(HaveLen(1))
		})
	})
})
