/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Pod Parser", func() {
	Describe("determineRevisionId", func() {
		It("pod has rev id label", func() {
			expectedRevId := uuid.NewString()
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.LabelRevisionId: expectedRevId,
					},
				},
			}
			revId := determineRevisionId(&pod)
			Expect(revId).To(Not(BeNil()))
			Expect(*revId).To(Equal(expectedRevId))
		})

		It("pod has invalid rev id label", func() {
			expectedRevId := "invalid"
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.LabelRevisionId: expectedRevId,
					},
				},
			}
			revId := determineRevisionId(&pod)
			Expect(revId).To(BeNil())
		})

		It("pod has doesn't have rev id label", func() {
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			}
			revId := determineRevisionId(&pod)
			Expect(revId).To(BeNil())
		})
	})

	Describe("getPodNodeAffinity", func() {
		var (
			required = &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "k1",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"v1", "v2"},
					}},
				}},
			}

			preferred = []corev1.PreferredSchedulingTerm{{
				Weight: 10,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "k2",
						Operator: corev1.NodeSelectorOpExists,
					}},
				},
			}}
		)

		It("should return nil for nil affinity", func() {
			result := getPodNodeAffinity(corev1.PodSpec{})
			Expect(result).To(BeNil())
		})

		It("should convert required node affinity", func() {
			spec := corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: required,
					},
				},
			}

			result := getPodNodeAffinity(spec)
			Expect(result).NotTo(BeNil())

			// Required
			required := result.Required.Get()
			Expect(required.NodeSelectorTerms).To(HaveLen(1))
			Expect(required.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
			me := required.NodeSelectorTerms[0].MatchExpressions[0]
			Expect(me.Key).To(Equal("k1"))
			Expect(string(me.Operator)).To(Equal(string(corev1.NodeSelectorOpIn)))
			Expect(me.Values).To(Equal([]string{"v1", "v2"}))

			// Preferred should be unset
			Expect(result.Preferred.IsSet()).To(BeFalse())
		})

		It("should convert preferred node affinity", func() {
			spec := corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: preferred,
					},
				},
			}

			result := getPodNodeAffinity(spec)
			Expect(result).NotTo(BeNil())

			// Preferred
			preferred := result.Preferred.Get()
			Expect(preferred.NodeSelectorTerms).To(HaveLen(1))
			term := preferred.NodeSelectorTerms[0]
			Expect(term.Weight).NotTo(BeNil())
			Expect(*term.Weight).To(Equal(float32(10)))
			pref := term.Preference.Get()
			Expect(pref.MatchExpressions).To(HaveLen(1))
			me := pref.MatchExpressions[0]
			Expect(me.Key).To(Equal("k2"))
			Expect(string(me.Operator)).To(Equal(string(corev1.NodeSelectorOpExists)))

			// Required should be unset
			Expect(result.Required.IsSet()).To(BeFalse())
		})

		It("should convert both required and preferred node affinity", func() {
			spec := corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution:  required,
						PreferredDuringSchedulingIgnoredDuringExecution: preferred,
					},
				},
			}
			result := getPodNodeAffinity(spec)
			Expect(result).NotTo(BeNil())
			Expect(result.Required).NotTo(BeNil())
			Expect(result.Preferred).NotTo(BeNil())
		})

		It("should handle empty node affinity (no terms)", func() {
			spec := corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution:  &corev1.NodeSelector{},
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{},
					},
				},
			}
			result := getPodNodeAffinity(spec)
			Expect(result).NotTo(BeNil())

			required := result.Required.Get()
			Expect(required.NodeSelectorTerms).To(BeEmpty())

			preferred := result.Preferred.Get()
			Expect(preferred.NodeSelectorTerms).To(BeEmpty())
		})
	})

	Describe("parsePodElements", func() {
		It("should parse all three element labels", func() {
			labels := map[string]string{
				"run.ai/element-1-id": "elem-aaa",
				"run.ai/element-2-id": "elem-bbb",
				"run.ai/element-3-id": "elem-ccc",
			}
			elements := parsePodElements(labels)
			Expect(elements).To(HaveLen(3))
			Expect(elements[0].ElementId).To(Equal("elem-aaa"))
			Expect(elements[0].HierarchyLevel).To(Equal(int32(1)))
			Expect(elements[1].ElementId).To(Equal("elem-bbb"))
			Expect(elements[1].HierarchyLevel).To(Equal(int32(2)))
			Expect(elements[2].ElementId).To(Equal("elem-ccc"))
			Expect(elements[2].HierarchyLevel).To(Equal(int32(3)))
		})

		It("should parse partial element labels (level 1 only)", func() {
			labels := map[string]string{
				"run.ai/element-1-id": "elem-aaa",
			}
			elements := parsePodElements(labels)
			Expect(elements).To(HaveLen(1))
			Expect(elements[0].ElementId).To(Equal("elem-aaa"))
			Expect(elements[0].HierarchyLevel).To(Equal(int32(1)))
		})

		It("should return nil for no element labels", func() {
			labels := map[string]string{
				"some-other-label": "value",
			}
			elements := parsePodElements(labels)
			Expect(elements).To(BeNil())
		})

		It("should return nil for nil labels", func() {
			elements := parsePodElements(nil)
			Expect(elements).To(BeNil())
		})

		It("should skip empty element label values", func() {
			labels := map[string]string{
				"run.ai/element-1-id": "elem-aaa",
				"run.ai/element-2-id": "",
				"run.ai/element-3-id": "elem-ccc",
			}
			elements := parsePodElements(labels)
			Expect(elements).To(HaveLen(2))
			Expect(elements[0].HierarchyLevel).To(Equal(int32(1)))
			Expect(elements[1].HierarchyLevel).To(Equal(int32(3)))
		})
	})

	Describe("parsePodTask", func() {
		It("should return 'leader' for LeaderWorkerSet pod with worker-index 0", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.LwsWorkerIndexLabelKey: "0",
					},
				},
			}
			task := parsePodTask(pod)
			Expect(task).NotTo(BeNil())
			Expect(*task).To(Equal(constants.PodTaskLeader))
		})

		It("should return 'worker' for LeaderWorkerSet pod with worker-index 1", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.LwsWorkerIndexLabelKey: "1",
					},
				},
			}
			task := parsePodTask(pod)
			Expect(task).NotTo(BeNil())
			Expect(*task).To(Equal(constants.PodTaskWorker))
		})

		It("should return 'worker' for LeaderWorkerSet pod with worker-index 5", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.LwsWorkerIndexLabelKey: "5",
					},
				},
			}
			task := parsePodTask(pod)
			Expect(task).NotTo(BeNil())
			Expect(*task).To(Equal(constants.PodTaskWorker))
		})

		It("should return nil for non-LeaderWorkerSet pod (no worker-index label)", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"some-other-label": "value",
					},
				},
			}
			task := parsePodTask(pod)
			Expect(task).To(BeNil())
		})

		It("should return nil for pod with no labels", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			}
			task := parsePodTask(pod)
			Expect(task).To(BeNil())
		})

		It("should return nil for pod with nil labels", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			}
			task := parsePodTask(pod)
			Expect(task).To(BeNil())
		})
	})

	Describe("parsePodResourceClaims", func() {
		It("should return nil for a pod without resource claims", func() {
			pod := &corev1.Pod{}
			Expect(parsePodResourceClaims(pod)).To(BeNil())
		})

		It("should report a template-generated claim with no allocated name until bound", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(1))
			Expect(claims[0].Name).To(Equal("gpu-1"))
			Expect(claims[0].ResourceClaimTemplateName.Get()).To(Equal(ptr.To("single-gpu")))
			Expect(claims[0].ResourceClaimName.Get()).To(BeNil())
			Expect(claims[0].AllocatedResourceClaimName.Get()).To(BeNil())
		})

		It("should take the allocated name of a template-generated claim from the pod status", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
					},
				},
				Status: corev1.PodStatus{
					ResourceClaimStatuses: []corev1.PodResourceClaimStatus{
						{Name: "gpu-1", ResourceClaimName: ptr.To("dra-gpu-test-gpu-1-mk4s8")},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(1))
			Expect(claims[0].ResourceClaimTemplateName.Get()).To(Equal(ptr.To("single-gpu")))
			Expect(claims[0].AllocatedResourceClaimName.Get()).To(Equal(ptr.To("dra-gpu-test-gpu-1-mk4s8")))
		})

		It("should use the spec name as the allocated name of a standalone claim", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "ib", ResourceClaimName: ptr.To("shared-infiniband")},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(1))
			Expect(claims[0].Name).To(Equal("ib"))
			Expect(claims[0].ResourceClaimTemplateName.Get()).To(BeNil())
			Expect(claims[0].ResourceClaimName.Get()).To(Equal(ptr.To("shared-infiniband")))
			Expect(claims[0].AllocatedResourceClaimName.Get()).To(Equal(ptr.To("shared-infiniband")))
		})

		It("should report mixed template and standalone claims in spec order", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
						{Name: "ib", ResourceClaimName: ptr.To("shared-infiniband")},
					},
				},
				Status: corev1.PodStatus{
					ResourceClaimStatuses: []corev1.PodResourceClaimStatus{
						{Name: "gpu-1", ResourceClaimName: ptr.To("pod-gpu-1-abcde")},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(2))
			Expect(claims[0].Name).To(Equal("gpu-1"))
			Expect(claims[0].AllocatedResourceClaimName.Get()).To(Equal(ptr.To("pod-gpu-1-abcde")))
			Expect(claims[1].Name).To(Equal("ib"))
			Expect(claims[1].AllocatedResourceClaimName.Get()).To(Equal(ptr.To("shared-infiniband")))
		})

		It("should report one entry per pod-level claim regardless of which containers consume it", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
					},
					Containers: []corev1.Container{
						{Name: "cuda", Resources: corev1.ResourceRequirements{Claims: []corev1.ResourceClaim{{Name: "gpu-1"}}}},
						{Name: "sidecar", Resources: corev1.ResourceRequirements{Claims: []corev1.ResourceClaim{{Name: "gpu-1"}}}},
						{Name: "ubuntu"},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(1))
			Expect(claims[0].Name).To(Equal("gpu-1"))
		})

		It("should ignore status entries that do not match a spec claim", func() {
			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					ResourceClaims: []corev1.PodResourceClaim{
						{Name: "gpu-1", ResourceClaimTemplateName: ptr.To("single-gpu")},
					},
				},
				Status: corev1.PodStatus{
					ResourceClaimStatuses: []corev1.PodResourceClaimStatus{
						{Name: "other", ResourceClaimName: ptr.To("stale-claim")},
					},
				},
			}

			claims := parsePodResourceClaims(pod)
			Expect(claims).To(HaveLen(1))
			Expect(claims[0].AllocatedResourceClaimName.Get()).To(BeNil())
		})
	})
})
