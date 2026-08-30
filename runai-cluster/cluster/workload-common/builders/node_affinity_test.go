package builders

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("NodeAffinity tests", func() {
	const (
		// affinity1
		key111      = "key111"
		operator111 = corev1.NodeSelectorOpExists
		value111    = "value111"
		key112      = "key112"
		operator112 = corev1.NodeSelectorOpExists
		value112    = "value112"

		key12      = "key12"
		operator12 = corev1.NodeSelectorOpDoesNotExist
		value12    = "value12"

		key13      = "key13"
		operator13 = corev1.NodeSelectorOpExists
		value13    = "value13"

		// affinity2
		key21      = "key21"
		operator21 = corev1.NodeSelectorOpExists
		value21    = "value21"

		key22      = "key22"
		operator22 = corev1.NodeSelectorOpDoesNotExist
		value22    = "value22"
	)

	var (
		affinity1, affinity2 *corev1.NodeAffinity
	)

	BeforeEach(func() {
		affinity1 = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							newMatchExpression(key111, operator111, value111),
							newMatchExpression(key112, operator112, value112),
						},
					},
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							newMatchExpression(key12, operator12, value12),
						},
					},
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							newMatchExpression(key13, operator13, value13),
						},
					},
				},
			},
		}

		affinity2 = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							newMatchExpression(key21, operator21, value21),
						},
					},
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							newMatchExpression(key22, operator22, value22),
						},
					},
				},
			},
		}
	})

	Context("Merge required node affinities", func() {
		It("full merge", func() {
			merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
			Expect(merged).NotTo(BeNil())

			required1 := affinity1.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			required2 := affinity2.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

			validateMergedTerm(required1[0], required2[0], merged.NodeSelectorTerms[0])
			validateMergedTerm(required1[0], required2[1], merged.NodeSelectorTerms[1])
			validateMergedTerm(required1[1], required2[0], merged.NodeSelectorTerms[2])
			validateMergedTerm(required1[1], required2[1], merged.NodeSelectorTerms[3])
			validateMergedTerm(required1[2], required2[0], merged.NodeSelectorTerms[4])
			validateMergedTerm(required1[2], required2[1], merged.NodeSelectorTerms[5])
			Expect(merged.NodeSelectorTerms).To(HaveLen(len(required1) * len(required2)))
		})

		It("remove duplicates in a term", func() {
			merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
			Expect(merged).NotTo(BeNil())

			// create a duplicate - add an expression from affinity1 to affinity2
			affinity2.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions = append(
				affinity2.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
				newMatchExpression(key111, operator111, value111),
			)

			mergedWithDuplicates := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
			Expect(mergedWithDuplicates).NotTo(BeNil())

			Expect(mergedWithDuplicates.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(3))
			Expect(merged.NodeSelectorTerms[0]).To(Equal(mergedWithDuplicates.NodeSelectorTerms[0]))
		})

		Context("no merge cases", func() {
			It("both nil", func() {
				merged := MergeRequiredNodeAffinities(nil, nil)
				Expect(merged).To(BeNil())
			})

			It("both empty", func() {
				affinity1.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, 0)
				affinity2.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, 0)
				merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
				Expect(merged.NodeSelectorTerms).To(HaveLen(0))
			})

			It("first is nil", func() {
				merged := MergeRequiredNodeAffinities(nil, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
				Expect(merged).To(Equal(affinity2.RequiredDuringSchedulingIgnoredDuringExecution))
			})

			It("first is empty", func() {
				affinity1.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, 0)
				merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
				Expect(merged).To(Equal(affinity2.RequiredDuringSchedulingIgnoredDuringExecution))
			})

			It("second nil", func() {
				merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, nil)
				Expect(merged).To(Equal(affinity1.RequiredDuringSchedulingIgnoredDuringExecution))
			})

			It("second is empty", func() {
				affinity2.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = make([]corev1.NodeSelectorTerm, 0)
				merged := MergeRequiredNodeAffinities(affinity1.RequiredDuringSchedulingIgnoredDuringExecution, affinity2.RequiredDuringSchedulingIgnoredDuringExecution)
				Expect(merged).To(Equal(affinity1.RequiredDuringSchedulingIgnoredDuringExecution))
			})
		})

		Context("mergeAffinityTerms", func() {
			It("remove duplicates", func() {
				term := affinity1.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]
				merged := mergeAffinityTerms(term, term)
				Expect(merged).NotTo(BeNil())
				Expect(merged).To(Equal(term))
			})
		})
	})
})

func validateMergedTerm(first, second, expected corev1.NodeSelectorTerm) {
	for _, expr := range first.MatchExpressions {
		Expect(expected.MatchExpressions).To(ContainElements(expr))
	}
	for _, expr := range second.MatchExpressions {
		Expect(expected.MatchExpressions).To(ContainElements(expr))
	}
}

func newMatchExpression(key string, operator corev1.NodeSelectorOperator, value string) corev1.NodeSelectorRequirement {
	return corev1.NodeSelectorRequirement{
		Key:      key,
		Operator: operator,
		Values:   []string{value},
	}
}
