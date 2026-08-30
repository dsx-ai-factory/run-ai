package affinity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func requirement(key string, op corev1.NodeSelectorOperator, values ...string) corev1.NodeSelectorRequirement {
	return corev1.NodeSelectorRequirement{Key: key, Operator: op, Values: values}
}

func selector(terms ...corev1.NodeSelectorTerm) *corev1.NodeSelector {
	return &corev1.NodeSelector{NodeSelectorTerms: terms}
}

func term(requirements ...corev1.NodeSelectorRequirement) corev1.NodeSelectorTerm {
	return corev1.NodeSelectorTerm{MatchExpressions: requirements}
}

var (
	gpuType  = requirement("gpu-type", corev1.NodeSelectorOpIn, "a100")
	zone     = requirement("zone", corev1.NodeSelectorOpIn, "us-east-1")
	excluded = requirement("kubernetes.io/hostname", corev1.NodeSelectorOpNotIn, "node-a")
)

var _ = Describe("MergeRequired", func() {
	It("returns nil when both sides are nil", func() {
		Expect(MergeRequired(nil, nil)).To(BeNil())
	})

	It("returns a copy of the non-empty side", func() {
		src := selector(term(excluded))

		merged := MergeRequired(selector(), src)

		Expect(merged).To(Equal(src))
		Expect(merged).NotTo(BeIdenticalTo(src), "the result must not alias its input")
	})

	It("ANDs every term of one side with every term of the other", func() {
		merged := MergeRequired(selector(term(gpuType), term(zone)), selector(term(excluded)))

		Expect(merged.NodeSelectorTerms).To(HaveLen(2))
		Expect(merged.NodeSelectorTerms[0].MatchExpressions).To(Equal([]corev1.NodeSelectorRequirement{gpuType, excluded}))
		Expect(merged.NodeSelectorTerms[1].MatchExpressions).To(Equal([]corev1.NodeSelectorRequirement{zone, excluded}))
	})

	// Drives the merge the way the framework's mutation loop does. A merge that
	// grows on re-application exhausts the budget, and the child is never applied.
	DescribeTable("settles within the framework's mutation-round budget",
		func(workload, instruction *corev1.NodeSelector) {
			const maxMutationRounds = 5

			current := workload
			var previous *corev1.NodeSelector
			rounds := 0
			for round := 1; round <= maxMutationRounds; round++ {
				previous = current
				current = MergeRequired(current, instruction)
				rounds = round
				if equality.Semantic.DeepEqual(previous, current) {
					break
				}
			}

			Expect(current).To(Equal(previous), "did not settle within %d rounds", maxMutationRounds)
			Expect(rounds).To(BeNumerically("<", maxMutationRounds))
			Expect(Implies(current, instruction)).To(BeTrue(), "the instruction must survive the merge")
		},
		Entry("single-term instruction onto a bare workload", selector(), selector(term(excluded))),
		Entry("single-term instruction onto a constrained workload", selector(term(gpuType)), selector(term(excluded))),
		// An OR: the cartesian merge squares the term count on re-application
		// unless it notices the work is already done.
		Entry("multi-term instruction onto a bare workload",
			selector(), selector(term(excluded), term(zone))),
		Entry("multi-term instruction onto a constrained workload",
			selector(term(gpuType)), selector(term(excluded), term(zone))),
		Entry("multi-term instruction onto a multi-term workload",
			selector(term(gpuType), term(requirement("pool", corev1.NodeSelectorOpExists))),
			selector(term(excluded), term(zone))),
		Entry("instruction the workload already states verbatim",
			selector(term(excluded)), selector(term(excluded))),
	)

	It("does not multiply terms when re-merging a multi-term instruction", func() {
		instruction := selector(term(excluded), term(zone))

		once := MergeRequired(selector(term(gpuType)), instruction)
		twice := MergeRequired(once, instruction)

		Expect(once.NodeSelectorTerms).To(HaveLen(2))
		Expect(twice).To(Equal(once))
	})

	// An empty term matches no nodes. Reading it as "satisfied by anything" would
	// skip the merge and silently discard the rest of the instruction.
	It("does not let an empty term suppress the rest of the selector", func() {
		child := selector(term(requirement("nvidia.com/gpu", corev1.NodeSelectorOpExists)))
		instruction := selector(corev1.NodeSelectorTerm{}, term(excluded))

		merged := MergeRequired(child, instruction)

		Expect(merged.NodeSelectorTerms).To(HaveLen(1))
		Expect(merged.NodeSelectorTerms[0].MatchExpressions).To(ContainElement(excluded))
	})

	// nil and zero terms place no constraint; a selector of only empty terms matches no nodes.
	DescribeTable("empty-shape semantics",
		func(first, second, expected *corev1.NodeSelector) {
			Expect(MergeRequired(first, second)).To(Equal(expected))
		},
		Entry("nil and nil place no constraint", nil, nil, nil),
		Entry("nil and zero terms place no constraint", nil, selector(), nil),
		Entry("zero terms on both sides place no constraint", selector(), selector(), nil),
		Entry("zero terms on the second side leave the first unchanged",
			selector(term(gpuType)), selector(), selector(term(gpuType))),
		Entry("a matches-nothing side wins over nil",
			selector(corev1.NodeSelectorTerm{}), nil, selector(corev1.NodeSelectorTerm{})),
		Entry("a matches-nothing side wins over zero terms",
			selector(corev1.NodeSelectorTerm{}), selector(), selector(corev1.NodeSelectorTerm{})),
		Entry("two matches-nothing sides stay matches-nothing",
			selector(corev1.NodeSelectorTerm{}), selector(corev1.NodeSelectorTerm{}), selector(corev1.NodeSelectorTerm{})),
		Entry("a matches-nothing instruction beats a real workload selector",
			selector(term(gpuType)), selector(corev1.NodeSelectorTerm{}), selector(corev1.NodeSelectorTerm{})),
		Entry("a matches-nothing workload selector beats a real instruction",
			selector(corev1.NodeSelectorTerm{}), selector(term(excluded)), selector(corev1.NodeSelectorTerm{})),
	)

	It("carries matchFields through the merge", func() {
		byName := corev1.NodeSelectorRequirement{
			Key:      "metadata.name",
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{"node-a"},
		}
		workload := &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchFields: []corev1.NodeSelectorRequirement{byName}}}}

		merged := MergeRequired(workload, selector(term(gpuType)))

		Expect(merged.NodeSelectorTerms[0].MatchFields).To(Equal([]corev1.NodeSelectorRequirement{byName}))
		Expect(merged.NodeSelectorTerms[0].MatchExpressions).To(Equal([]corev1.NodeSelectorRequirement{gpuType}))
	})
})

var _ = Describe("Implies", func() {
	It("holds when every term carries one of the required terms", func() {
		Expect(Implies(selector(term(gpuType, excluded), term(zone, excluded)), selector(term(excluded)))).To(BeTrue())
	})

	It("fails when one term is missing the requirement", func() {
		Expect(Implies(selector(term(gpuType, excluded), term(zone)), selector(term(excluded)))).To(BeFalse())
	})

	It("holds against a multi-term selector when each term satisfies a different branch", func() {
		Expect(Implies(selector(term(gpuType, excluded), term(gpuType, zone)), selector(term(excluded), term(zone)))).To(BeTrue())
	})

	It("is false for an empty first selector and true for an empty second", func() {
		Expect(Implies(selector(), selector(term(excluded)))).To(BeFalse())
		Expect(Implies(selector(term(excluded)), selector())).To(BeTrue())
	})

	It("is false against an empty term, which matches no nodes", func() {
		Expect(Implies(selector(term(gpuType)), selector(corev1.NodeSelectorTerm{}))).To(BeFalse())
	})

	It("ignores value ordering", func() {
		have := requirement("k", corev1.NodeSelectorOpNotIn, "b", "a")
		want := requirement("k", corev1.NodeSelectorOpNotIn, "a", "b")

		Expect(Implies(selector(term(have)), selector(term(want)))).To(BeTrue())
	})
})

var _ = Describe("MergeRequired requirement de-duplication", func() {
	It("keeps requirements that share a key but differ in values", func() {
		firstExclusion := requirement("kubernetes.io/hostname", corev1.NodeSelectorOpNotIn, "node-a")
		secondExclusion := requirement("kubernetes.io/hostname", corev1.NodeSelectorOpNotIn, "node-b")

		merged := MergeRequired(selector(term(firstExclusion)), selector(term(secondExclusion)))

		Expect(merged.NodeSelectorTerms[0].MatchExpressions).To(Equal(
			[]corev1.NodeSelectorRequirement{firstExclusion, secondExclusion}))
	})

	It("treats value order as insignificant when de-duplicating", func() {
		merged := MergeRequired(
			selector(term(requirement("k", corev1.NodeSelectorOpNotIn, "a", "b"))),
			selector(term(requirement("k", corev1.NodeSelectorOpNotIn, "b", "a"))),
		)

		Expect(merged.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
	})
})

var _ = Describe("MergeNodeAffinity", func() {
	preferred := func(weight int32, key string) corev1.PreferredSchedulingTerm {
		return corev1.PreferredSchedulingTerm{
			Weight:     weight,
			Preference: term(requirement(key, corev1.NodeSelectorOpExists)),
		}
	}

	It("merges required and preferred terms", func() {
		dst := &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  selector(term(gpuType)),
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(10, "fast")},
		}
		src := &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  selector(term(excluded)),
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(20, "spot")},
		}

		MergeNodeAffinity(dst, src)

		Expect(dst.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).
			To(Equal([]corev1.NodeSelectorRequirement{gpuType, excluded}))
		Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(2))
	})

	It("does not duplicate a preferred term that is already present", func() {
		dst := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(10, "fast")},
		}
		src := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(10, "fast")},
		}

		MergeNodeAffinity(dst, src)

		Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
	})

	It("does not duplicate a preferred term whose values differ only in order", func() {
		zoneIn := func(values ...string) corev1.PreferredSchedulingTerm {
			return corev1.PreferredSchedulingTerm{
				Weight:     50,
				Preference: term(requirement("zone", corev1.NodeSelectorOpIn, values...)),
			}
		}
		dst := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{zoneIn("b", "a")},
		}
		src := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{zoneIn("a", "b")},
		}

		MergeNodeAffinity(dst, src)

		Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
	})

	It("keeps a preferred term that differs only in weight", func() {
		dst := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(10, "fast")},
		}
		src := &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{preferred(20, "fast")},
		}

		MergeNodeAffinity(dst, src)

		Expect(dst.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(2))
	})
})

var _ = Describe("NormalizeNodeAffinity", func() {
	It("sorts and de-duplicates requirement values", func() {
		affinity := &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: selector(
				term(requirement("kubernetes.io/hostname", corev1.NodeSelectorOpNotIn, "node-c", "node-a", "node-c"))),
		}

		NormalizeNodeAffinity(affinity)

		Expect(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values).
			To(Equal([]string{"node-a", "node-c"}))
	})

	It("tolerates a nil affinity", func() {
		Expect(func() { NormalizeNodeAffinity(nil) }).NotTo(Panic())
	})
})
