package converter

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// UseKaiNodePools is a compile-time const, so only one branch of the node pool lookup in
// convertNodeAffinityToNodePoolNames is reachable in any given build. Following the
// nodepool-controller's project-reference specs, every fixture here is written to BOTH
// CRDs with identical content, so these assertions hold — and keep holding — whichever
// value the const has.
// Delete the legacy fixtures alongside the else branch.
var _ = Describe("convertNodeAffinityToNodePoolNames", func() {
	const (
		affinityLabelKey     = "affinity/key"
		affinityLabelValue   = "affinity-value"
		nodepoolName         = "np-affinity"
		deletingNodepoolName = "np-deleting"
		deletingLabelValue   = "deleting-value"

		// The caller-supplied vocabulary; see the identifiers built in BeforeEach.
		assignmentLabelKey  = "example.com/node-pool"
		defaultNodepoolName = "example-default"
	)

	var (
		ctx          context.Context
		readerClient client.Client
		identifiers  NodePoolIdentifiers
	)

	affinityFor := func(key, value string) *corev1.Affinity {
		return &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      key,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{value},
						}},
					}},
				},
			},
		}
	}

	BeforeEach(func() {
		ctx = context.Background()

		// Deliberately non-production values. This package takes its whole node-pool
		// vocabulary from the caller — pod-group-assigner passes flag-driven config,
		// cluster-sync and workload-exporter pass run:ai constants — so a production key
		// hardcoded here would hide a regression that reintroduced one in the converter.
		identifiers = NodePoolIdentifiers{
			NodePoolAssignmentLabelKey: assignmentLabelKey,
			DefaultNodepoolName:        defaultNodepoolName,
			UnexistingNodepoolSentinel: "example-unexisting-node-pool",
			AnnotationNodepoolsKey:     "example-nodepools",
		}

		scheme := runtime.NewScheme()
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())

		// Same names and labels on both CRDs, so the expected result does not depend on
		// which one the const selects.
		readerClient = fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(
			&runv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: nodepoolName},
				Spec:       runv1alpha1.NodePoolSpec{LabelKey: affinityLabelKey, LabelValue: affinityLabelValue},
			},
			&runv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: deletingNodepoolName},
				Spec:       runv1alpha1.NodePoolSpec{LabelKey: affinityLabelKey, LabelValue: deletingLabelValue},
				Status:     runv1alpha1.NodePoolStatus{Phase: runv1alpha1.NodePoolDeleting},
			},
			&kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: nodepoolName},
				Spec:       kaiv1alpha1.NodePoolSpec{LabelKey: affinityLabelKey, LabelValue: affinityLabelValue},
			},
			&kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: deletingNodepoolName},
				Spec:       kaiv1alpha1.NodePoolSpec{LabelKey: affinityLabelKey, LabelValue: deletingLabelValue},
				Status:     kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolDeleting},
			},
		).Build()
	})

	It("resolves a matching node pool by label key and value", func() {
		result, err := convertNodeAffinityToNodePoolNames(
			ctx, affinityFor(affinityLabelKey, affinityLabelValue), identifiers, readerClient)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]string{nodepoolName}))
	})

	// A pool being deleted must not be offered for scheduling, even though its labels
	// match the affinity term.
	It("skips a node pool that is being deleted", func() {
		result, err := convertNodeAffinityToNodePoolNames(
			ctx, affinityFor(affinityLabelKey, deletingLabelValue), identifiers, readerClient)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	It("returns nothing for a term that matches no node pool", func() {
		result, err := convertNodeAffinityToNodePoolNames(
			ctx, affinityFor("other/key", "other-value"), identifiers, readerClient)

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeEmpty())
	})

	// The default nodepool is the nodes belonging to no other nodepool, so it is requested by the
	// absence of the assignment label rather than by a match on a pool's own labels.
	Describe("the default node pool", func() {
		doesNotExistOn := func(key string) *corev1.Affinity {
			return &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{{
							MatchExpressions: []corev1.NodeSelectorRequirement{{
								Key:      key,
								Operator: corev1.NodeSelectorOpDoesNotExist,
							}},
						}},
					},
				},
			}
		}

		It("is resolved from a DoesNotExist term on the caller's assignment label key", func() {
			result, err := convertNodeAffinityToNodePoolNames(
				ctx, doesNotExistOn(assignmentLabelKey), identifiers, readerClient)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{defaultNodepoolName}))
		})

		// Guards against the converter reintroducing a hardcoded run:ai key: with one, the
		// spec above would stop resolving and this one would start.
		It("is not resolved from a DoesNotExist term on any other key", func() {
			result, err := convertNodeAffinityToNodePoolNames(
				ctx, doesNotExistOn("runai/node-pool"), identifiers, readerClient)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	DescribeTable("short-circuits before reading any node pool",
		func(affinity *corev1.Affinity) {
			result, err := convertNodeAffinityToNodePoolNames(ctx, affinity, identifiers, readerClient)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEmpty())
		},
		Entry("nil affinity", nil),
		Entry("no node affinity", &corev1.Affinity{}),
		Entry("no required node selector", &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}),
		Entry("no node selector terms", &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
			},
		}),
	)

	// The listing itself is the only part that differs between the two CRDs, and both
	// implementations are covered directly in the utils package's specs.
	It("errors when the selected CRD has no node pools at all", func() {
		emptyScheme := runtime.NewScheme()
		Expect(runv1alpha1.AddToScheme(emptyScheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(emptyScheme)).To(Succeed())
		emptyClient := fakeclient.NewClientBuilder().WithScheme(emptyScheme).Build()

		_, err := convertNodeAffinityToNodePoolNames(
			ctx, affinityFor(affinityLabelKey, affinityLabelValue), identifiers, emptyClient)

		Expect(err).To(MatchError(ContainSubstring("didn't find any node pools")))
	})
})
