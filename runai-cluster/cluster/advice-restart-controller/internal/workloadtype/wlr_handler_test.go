package workloadtype

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

func newWLR(nodeAffinity *corev1.NodeAffinity) *unstructured.Unstructured {
	GinkgoHelper()
	wlr := &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GVKWorkloadRunner.GroupVersion().String(),
			Kind:       GVKWorkloadRunner.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "wlr", Namespace: "ns"},
		Status:     dromosv1alpha1.WorkloadRunnerStatus{Phase: "Running"},
	}
	if nodeAffinity != nil {
		wlr.Spec.Instructions = &dromosv1alpha1.Instructions{
			NodeAffinity: &dromosv1alpha1.NodeAffinityConfig{NodeAffinity: nodeAffinity},
		}
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(wlr)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{Object: obj}
}

var _ = Describe("WorkloadRunnerHandler", func() {
	Describe("NewWorkloadRunnerHandler", func() {
		It("errors on an instruction with a wrong-typed declared field", func() {
			wlr := newWLR(nil)
			path := []string{"spec", "instructions", "nodeAffinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution"}
			err := unstructured.SetNestedField(wlr.Object, "not-an-object", path...)
			Expect(err).NotTo(HaveOccurred())
			_, err = NewWorkloadRunnerHandler(wlr)
			Expect(err).To(HaveOccurred())
		})

		It("errors on a non-map instruction", func() {
			wlr := newWLR(nil)
			path := []string{"spec", "instructions", "nodeAffinity", "nodeAffinity"}
			err := unstructured.SetNestedField(wlr.Object, "garbage", path...)
			Expect(err).NotTo(HaveOccurred())
			_, err = NewWorkloadRunnerHandler(wlr)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetPhase", func() {
		It("reads status.phase", func() {
			wlr := newWLR(nil)
			handler, err := NewWorkloadRunnerHandler(wlr)
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetPhase()).To(Equal("Running"))
		})
	})

	DescribeTable("BuildSpecOverrides",
		func(existingAffinity *corev1.NodeAffinity, exclusions, wantExcluded []string) {
			wlr := newWLR(existingAffinity)
			handler, err := NewWorkloadRunnerHandler(wlr)
			Expect(err).NotTo(HaveOccurred())

			overrides, err := handler.BuildSpecOverrides(exclusions)
			Expect(err).NotTo(HaveOccurred())

			if wantExcluded == nil {
				Expect(overrides).To(BeNil())
				return
			}

			Expect(overrides).To(HaveLen(1))
			Expect(overrides[0].Op).To(Equal(runv1alpha1.JSONPatchOperationAdd))
			Expect(overrides[0].Path).To(Equal("/spec/instructions/nodeAffinity"))

			wrapped := map[string]*corev1.NodeAffinity{}
			Expect(json.Unmarshal(overrides[0].Value.Raw, &wrapped)).To(Succeed())
			Expect(wrapped).To(HaveLen(1))
			Expect(wrapped["nodeAffinity"]).NotTo(BeNil())

			required := wrapped["nodeAffinity"].RequiredDuringSchedulingIgnoredDuringExecution
			Expect(required).NotTo(BeNil())
			Expect(required.NodeSelectorTerms).To(HaveLen(1))
			Expect(required.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))

			expr := required.NodeSelectorTerms[0].MatchExpressions[0]
			Expect(expr.Key).To(Equal(corev1.LabelHostname))
			Expect(expr.Operator).To(Equal(corev1.NodeSelectorOpNotIn))
			Expect(expr.Values).To(Equal(wantExcluded))
		},

		Entry("no exclusions: no override",
			&corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      corev1.LabelHostname,
							Operator: corev1.NodeSelectorOpNotIn,
							Values:   []string{"node-a"},
						}},
					}},
				},
			},
			nil, nil),

		Entry("instruction absent: just the exclusions",
			nil,
			[]string{"node-b", "node-a"}, []string{"node-a", "node-b"}),

		Entry("existing exclusions: union, sorted and deduped",
			&corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      corev1.LabelHostname,
							Operator: corev1.NodeSelectorOpNotIn,
							Values:   []string{"node-b", "node-a"},
						}},
					}},
				},
			},
			[]string{"node-c", "node-a"}, []string{"node-a", "node-b", "node-c"}),

		Entry("prior instruction: dropped, hostname exclusions kept",
			&corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      corev1.LabelHostname,
								Operator: corev1.NodeSelectorOpNotIn,
								Values:   []string{"node-a"},
							},
							{
								Key:      "nvidia.com/gpu.product",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"H100"},
							},
						},
					}},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
					Weight: 1,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "zone",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"a"},
						}},
					},
				}},
			},
			[]string{"node-b"}, []string{"node-a", "node-b"}),
	)
})
