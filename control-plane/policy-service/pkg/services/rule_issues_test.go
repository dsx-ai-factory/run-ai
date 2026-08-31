/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// Pure unit tests for the ruleIssues collector. validateRules-integration
// tests (multi-rule accumulation, full() short-circuiting via the rule loop,
// breakdown-generator + validateRules sharing) live in rules_validator_test.go.
var _ = Describe("ruleIssues", func() {
	Describe("newRuleIssues", func() {
		It("returns an empty, not-full collector regardless of maxErrors", func() {
			for _, max := range []int{0, 1, 5} {
				r := newRuleIssues(max)
				Expect(r.all()).To(BeEmpty())
				Expect(r.full()).To(BeFalse())
				Expect(r.toError()).To(BeNil())
			}
		})
	})

	Describe("add", func() {
		It("ignores nil errors", func() {
			r := newRuleIssues(0)
			r.add(nil)
			Expect(r.all()).To(BeEmpty())
		})

		It("appends non-nil errors in insertion order", func() {
			r := newRuleIssues(0)
			r.add(api.NewValidationError("first"))
			r.add(api.NewValidationError("second"))
			Expect(r.all()).To(HaveLen(2))
			Expect(r.all()[0].Message).To(Equal("first"))
			Expect(r.all()[1].Message).To(Equal("second"))
		})
	})

	Describe("addf", func() {
		It("formats the message using fmt.Sprintf and appends a validation error", func() {
			r := newRuleIssues(0)
			r.addf("rule[%d] (field=%s): %s", 3, "image", "broken")
			Expect(r.all()).To(HaveLen(1))
			Expect(r.all()[0].Message).To(Equal("rule[3] (field=image): broken"))
		})
	})

	Describe("addWithDetails", func() {
		It("appends an error carrying both Message and Details", func() {
			r := newRuleIssues(0)
			r.addWithDetails("the message", "the details body")
			Expect(r.all()).To(HaveLen(1))
			Expect(r.all()[0].Message).To(Equal("the message"))
			Expect(r.all()[0].Details).To(Equal("the details body"))
		})
	})

	Describe("full / maxErrors cap", func() {
		It("with maxErrors=0 is never full (unlimited)", func() {
			r := newRuleIssues(0)
			for i := 0; i < 100; i++ {
				r.addf("issue %d", i)
			}
			Expect(r.full()).To(BeFalse())
			Expect(r.all()).To(HaveLen(100))
		})

		It("with maxErrors > 0 becomes full once exactly that many issues have been collected", func() {
			r := newRuleIssues(2)
			Expect(r.full()).To(BeFalse())
			r.addf("first")
			Expect(r.full()).To(BeFalse())
			r.addf("second")
			Expect(r.full()).To(BeTrue())
		})

		It("treats every add* call as a no-op once the collector is full", func() {
			r := newRuleIssues(2)
			r.addf("first")
			r.addf("second")

			r.addf("third — dropped")
			r.add(api.NewValidationError("fourth — dropped"))
			r.addWithDetails("fifth — dropped", "details")

			Expect(r.all()).To(HaveLen(2))
			Expect(r.all()[0].Message).To(Equal("first"))
			Expect(r.all()[1].Message).To(Equal("second"))
		})

		It("with maxErrors=1 retains only the first issue (legacy fail-fast behavior)", func() {
			r := newRuleIssues(1)
			r.addf("first")
			r.addf("second")
			Expect(r.full()).To(BeTrue())
			Expect(r.all()).To(HaveLen(1))
			Expect(r.toError().Message).To(Equal("first"))
		})
	})

	Describe("toError", func() {
		It("returns nil when no issues have been collected", func() {
			r := newRuleIssues(0)
			Expect(r.toError()).To(BeNil())
		})

		It("returns the only issue verbatim (Message + Details preserved)", func() {
			r := newRuleIssues(0)
			r.addWithDetails("only", "with-details")
			err := r.toError()
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(Equal("only"))
			Expect(err.Details).To(Equal("with-details"))
		})

		It("returns the first issue when many are collected, with the rest reachable via all()", func() {
			r := newRuleIssues(0)
			for i := 0; i < 5; i++ {
				r.addf("issue %d", i)
			}
			Expect(r.toError().Message).To(Equal("issue 0"))
			Expect(r.all()).To(HaveLen(5))
			Expect(r.all()[4].Message).To(Equal("issue 4"))
		})
	})

	Describe("all", func() {
		It("returns the issues in insertion order", func() {
			r := newRuleIssues(0)
			r.addf("a")
			r.add(api.NewValidationError("b"))
			r.addWithDetails("c", "c-details")
			msgs := []string{r.all()[0].Message, r.all()[1].Message, r.all()[2].Message}
			Expect(msgs).To(Equal([]string{"a", "b", "c"}))
		})

		It("returns an empty slice when nothing has been collected", func() {
			Expect(newRuleIssues(0).all()).To(BeEmpty())
		})
	})
})
