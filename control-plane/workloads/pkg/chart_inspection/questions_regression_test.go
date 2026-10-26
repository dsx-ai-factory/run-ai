// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Declaration normalization contract", func() {
	It("preserves warnings before reporting the first validation failure", func() {
		result := inspectQuestions(`questions:
- variable: setting
  type: custom-widget
  group: 123
  min: invalid
`, "{}")
		Expect(result.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE))
		Expect(result.Questions).To(BeEmpty())
		Expect(result.Groups).To(BeEmpty())
		Expect(result.Warnings).To(HaveLen(3))
		Expect(result.Warnings[0].Code).To(Equal("UNKNOWN_TYPE"))
		Expect(result.Warnings[1].Code).To(Equal("LABEL_DEFAULTED"))
		Expect(result.Warnings[2].Code).To(Equal("UNUSABLE_QUESTIONS"))
		Expect(result.Warnings[2].Message).To(ContainSubstring("questions[0] group must be text"))
	})

	It("distinguishes omitted children from an explicit empty list", func() {
		result := inspectQuestions(`questions:
- {variable: first, type: boolean, label: First}
- {variable: second, type: boolean, label: Second, subquestions: [], show_subquestions_if: false}
`, "{}")
		Expect(result.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE))
		Expect(result.Questions[0]).NotTo(HaveKey("subquestions"))
		Expect(result.Questions[1]).To(HaveKeyWithValue("subquestions", []any{}))
		Expect(result.Questions[1]).To(HaveKeyWithValue("show_subquestion_if", false))
		Expect(result.Questions[1]).NotTo(HaveKey("show_subquestions_if"))
	})

	DescribeTable("visibility dependencies", func(body string, usable bool) {
		result := inspectQuestions(body, "{}")
		if usable {
			Expect(result.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE))
		} else {
			Expect(result.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE))
			Expect(result.Warnings[len(result.Warnings)-1].Message).To(ContainSubstring("question visibility conditions contain a cycle"))
		}
	},
		Entry("canonical self-reference", `questions: [{variable: a.b, type: boolean, show_if: 'a["b"]=true'}]`, false),
		Entry("cycle disconnected from the first field", `questions: [{variable: first, type: string}, {variable: x, type: boolean, show_if: 'y=true'}, {variable: y, type: boolean, show_if: 'x=true'}]`, false),
		Entry("forward reference without a cycle", `questions: [{variable: x, type: boolean, show_if: 'y=true'}, {variable: y, type: boolean}]`, true),
		Entry("parent depends on its child", `questions: [{variable: parent, type: boolean, show_if: 'child=true', subquestions: [{variable: child, type: boolean}]}]`, false),
	)
})
