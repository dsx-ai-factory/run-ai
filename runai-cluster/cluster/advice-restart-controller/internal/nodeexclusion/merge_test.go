/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeexclusion

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("MergeNodeExclusions", func() {
	It("should create new affinity when no existing affinity", func() {
		result := MergeNodeExclusions(nil, []string{"node-1"})

		Expect(result).NotTo(BeNil())
		terms := result.NodeAffinityRequired.NodeSelectorTerms
		Expect(terms).To(HaveLen(1))
		Expect(terms[0].MatchExpressions).To(HaveLen(1))
		expectExpression(terms[0].MatchExpressions[0], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-1"})
	})

	It("should preserve existing GPU affinity and append exclusion", func() {
		existing := &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{
				NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "nvidia.com/gpu.product",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"A100-SXM4-80GB"},
							},
						},
					},
				},
			},
		}

		result := MergeNodeExclusions(existing, []string{"node-2"})

		terms := result.NodeAffinityRequired.NodeSelectorTerms
		Expect(terms).To(HaveLen(1))
		Expect(terms[0].MatchExpressions).To(HaveLen(2))
		expectExpression(terms[0].MatchExpressions[0], "nvidia.com/gpu.product", corev1.NodeSelectorOpIn, []string{"A100-SXM4-80GB"})
		expectExpression(terms[0].MatchExpressions[1], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-2"})
	})

	It("should accumulate exclusions across restarts", func() {
		existing := &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{
				NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      hostnameKey,
								Operator: corev1.NodeSelectorOpNotIn,
								Values:   []string{"node-1"},
							},
						},
					},
				},
			},
		}

		result := MergeNodeExclusions(existing, []string{"node-2"})

		expr := result.NodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions
		Expect(expr).To(HaveLen(1))
		expectExpression(expr[0], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-1", "node-2"})
	})

	It("should deduplicate node names", func() {
		existing := &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{
				NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      hostnameKey,
								Operator: corev1.NodeSelectorOpNotIn,
								Values:   []string{"node-1"},
							},
						},
					},
				},
			},
		}

		result := MergeNodeExclusions(existing, []string{"node-1", "node-2"})

		expr := result.NodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions
		expectExpression(expr[0], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-1", "node-2"})
	})

	It("should append exclusion to each node selector term", func() {
		existing := &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{
				NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1a"}},
						},
					},
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1b"}},
						},
					},
				},
			},
		}

		result := MergeNodeExclusions(existing, []string{"node-1"})

		terms := result.NodeAffinityRequired.NodeSelectorTerms
		Expect(terms).To(HaveLen(2))
		for _, term := range terms {
			Expect(term.MatchExpressions).To(HaveLen(2))
			expectExpression(term.MatchExpressions[1], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-1"})
		}
	})

	It("should return existing unchanged when exclusions are empty", func() {
		existing := &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{
				NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"us-east-1a"}},
						},
					},
				},
			},
		}

		result := MergeNodeExclusions(existing, []string{})

		Expect(result.NodeAffinityRequired.NodeSelectorTerms).To(HaveLen(1))
		Expect(result.NodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
	})

	It("should return nil for nil existing and empty exclusions", func() {
		result := MergeNodeExclusions(nil, []string{})
		Expect(result).To(BeNil())
	})

	It("should create affinity when existing has nil NodeAffinityRequired", func() {
		existing := &field_types.NodeAffinityFields{}

		result := MergeNodeExclusions(existing, []string{"node-1"})

		Expect(result).NotTo(BeNil())
		terms := result.NodeAffinityRequired.NodeSelectorTerms
		Expect(terms).To(HaveLen(1))
		expectExpression(terms[0].MatchExpressions[0], hostnameKey, corev1.NodeSelectorOpNotIn, []string{"node-1"})
	})
})

var _ = Describe("unionStrings", func() {
	DescribeTable("should compute sorted union",
		func(a, b, want []string) {
			Expect(unionStrings(a, b)).To(Equal(want))
		},
		Entry("both empty", nil, nil, []string{}),
		Entry("a only", []string{"b", "a"}, nil, []string{"a", "b"}),
		Entry("b only", nil, []string{"c", "a"}, []string{"a", "c"}),
		Entry("overlap", []string{"a", "b"}, []string{"b", "c"}, []string{"a", "b", "c"}),
		Entry("duplicates in input", []string{"a", "a"}, []string{"a"}, []string{"a"}),
	)
})

func expectExpression(expr corev1.NodeSelectorRequirement, key string, op corev1.NodeSelectorOperator, values []string) {
	GinkgoHelper()
	Expect(expr.Key).To(Equal(key))
	Expect(expr.Operator).To(Equal(op))
	Expect(expr.Values).To(Equal(values))
}
