/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structure

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	ewv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

func TestStructure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Structure Suite")
}

// ============================================================================
// Helper builders for tests
// ============================================================================

func makeKarta(rootName string, children []v1alpha1.ComponentDefinition) *v1alpha1.Karta {
	return &v1alpha1.Karta{
		Spec: v1alpha1.KartaSpec{
			StructureDefinition: v1alpha1.StructureDefinition{
				RootComponent: v1alpha1.ComponentDefinition{
					Name: rootName,
				},
				ChildComponents: children,
			},
		},
	}
}

func makePod(uid, name string, labels map[string]string) corev1.Pod {
	return makePodWithAnnotations(uid, name, labels, map[string]string{})
}

func makePodWithAnnotations(uid, name string, labels, annotations map[string]string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:         types.UID(uid),
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
}

// labelPods applies element labels from computed addresses to the corresponding pods.
// Used in lifecycle tests to simulate the label-after-build step.
func labelPods(pods []corev1.Pod, addresses []*PodAddress) []corev1.Pod {
	labeled := make([]corev1.Pod, len(pods))
	copy(labeled, pods)
	for _, addr := range addresses {
		labels := GeneratePodElementLabels(addr.ElementsPath)
		for i := range labeled {
			if string(labeled[i].UID) == addr.PodID {
				if labeled[i].Labels == nil {
					labeled[i].Labels = map[string]string{}
				}
				for k, v := range labels {
					labeled[i].Labels[k] = v
				}
			}
		}
	}
	return labeled
}

// findByK8sUID finds a StructureElement by its K8sRef.UID in an element map.
func findByK8sUID(elements map[string]*ewv1alpha1.StructureElement, uid string) *ewv1alpha1.StructureElement {
	for _, e := range elements {
		if e.K8sRef != nil && string(e.K8sRef.UID) == uid {
			return e
		}
	}
	return nil
}

// ============================================================================
// GetPodElementsPath Tests
// ============================================================================

var _ = Describe("GetPodElementsPath", func() {

	Context("Strategy B: Kind in owner chain", func() {
		It("should match pod to element via owner chain Kind", func() {
			karta := makeKarta("lws", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
				},
			})

			pod := makePod("pod-1", "worker-0", nil)
			ownerChain := OwnerChain{
				{UID: "rs-uid", Name: "worker-rs", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
				{UID: "sts-uid-123", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				{UID: "lws-uid", Name: "my-lws", Kind: "LeaderWorkerSet", Group: "leaderworkerset.x-k8s.io", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.PodID).To(Equal("pod-1"))
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].ID).NotTo(BeEmpty())
			Expect(addr.ElementsPath[0].Name).To(Equal("worker"))
			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid-123")))
			Expect(addr.ElementsPath[0].K8sRef.Name).To(Equal("my-sts"))
			Expect(addr.ElementsPath[0].Level).To(Equal(1))
		})
	})

	Context("Strategy A: ComponentTypeSelector matching", func() {
		It("should match pod using label selector with value", func() {
			workerValue := "worker"
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["app.kubernetes.io/component"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			pod := makePod("pod-1", "worker-pod", map[string]string{
				"app.kubernetes.io/component": "worker",
			})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].Name).To(Equal("worker"))
			Expect(addr.ElementsPath[0].K8sRef).To(BeNil())
			Expect(addr.ElementsPath[0].Level).To(Equal(1))
			Expect(addr.ElementsPath[0].ID).NotTo(BeEmpty())
		})

		It("should not match pod when selector value doesn't match (strategy A)", func() {
			workerValue := "worker"
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["app.kubernetes.io/component"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			pod := makePod("pod-1", "other-pod", map[string]string{
				"app.kubernetes.io/component": "scheduler",
			})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(BeEmpty())
		})
	})

	Context("Strategy C: Logical element", func() {
		It("should always match pod to logical element (no Kind, no selector)", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
				},
			})

			pod := makePod("pod-1", "any-pod", nil)

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].Name).To(Equal("group"))
			Expect(addr.ElementsPath[0].K8sRef).To(BeNil())
		})
	})

	Context("ComponentInstanceSelector", func() {
		It("should produce different instances for different pod label values (Dynamo-style)", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "service",
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["nvidia.com/dynamo-component"]`,
						},
					},
				},
			})

			podFrontend := makePod("pod-fe", "frontend-pod", map[string]string{
				"nvidia.com/dynamo-component": "Frontend",
			})
			podPrefill := makePod("pod-pw", "prefill-pod", map[string]string{
				"nvidia.com/dynamo-component": "PrefillWorker",
			})
			podDecode := makePod("pod-dw", "decode-pod", map[string]string{
				"nvidia.com/dynamo-component": "DecodeWorker",
			})

			addrFe := GetPodElementsPath(context.Background(), &podFrontend, karta, nil, nil)
			addrPw := GetPodElementsPath(context.Background(), &podPrefill, karta, nil, nil)
			addrDw := GetPodElementsPath(context.Background(), &podDecode, karta, nil, nil)

			Expect(addrFe.ElementsPath).To(HaveLen(1))
			Expect(addrPw.ElementsPath).To(HaveLen(1))
			Expect(addrDw.ElementsPath).To(HaveLen(1))

			Expect(addrFe.ElementsPath[0].InstanceKey).To(Equal(ptr.To("Frontend")))
			Expect(addrPw.ElementsPath[0].InstanceKey).To(Equal(ptr.To("PrefillWorker")))
			Expect(addrDw.ElementsPath[0].InstanceKey).To(Equal(ptr.To("DecodeWorker")))

			// All should have different element IDs (UUIDs)
			Expect(addrFe.ElementsPath[0].ID).NotTo(Equal(addrPw.ElementsPath[0].ID))
			Expect(addrFe.ElementsPath[0].ID).NotTo(Equal(addrDw.ElementsPath[0].ID))
			Expect(addrPw.ElementsPath[0].ID).NotTo(Equal(addrDw.ElementsPath[0].ID))

			// All should have different display names
			Expect(addrFe.ElementsPath[0].Name).To(Equal("service:frontend"))
			Expect(addrPw.ElementsPath[0].Name).To(Equal("service:prefillworker"))
			Expect(addrDw.ElementsPath[0].Name).To(Equal("service:decodeworker"))

		})
	})

	Context("UUID reuse via Name", func() {
		It("should reuse existing element UUID when Name matches", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{Name: "group"},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"existing-uuid-123": {
					ID:   "existing-uuid-123",
					Name: "group",
				},
			}

			pod := makePod("pod-1", "my-pod", nil)

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, existing)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].ID).To(Equal("existing-uuid-123"))
		})
	})

	Context("Empty owner chain", func() {
		It("should not match ObjectBacked element with empty owner chain", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
				},
			})

			pod := makePod("pod-1", "my-pod", nil)

			addr := GetPodElementsPath(context.Background(), &pod, karta, OwnerChain{}, nil)
			Expect(addr.ElementsPath).To(BeEmpty())
		})
	})

	Context("Multi-level hierarchy", func() {
		It("should produce multi-level path for nested elements", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
				},
			})

			pod := makePod("pod-1", "worker-0", nil)
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(2))

			// Level 1: ObjectBacked group
			Expect(addr.ElementsPath[0].Name).To(Equal("group"))
			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
			Expect(addr.ElementsPath[0].K8sRef.Name).To(Equal("my-sts"))
			Expect(addr.ElementsPath[0].ID).NotTo(BeEmpty())
			Expect(addr.ElementsPath[0].Level).To(Equal(1))

			// Level 2: Logical worker
			Expect(addr.ElementsPath[1].Name).To(Equal("worker"))
			Expect(addr.ElementsPath[1].K8sRef).To(BeNil())
			Expect(addr.ElementsPath[1].Level).To(Equal(2))
		})
	})

	Context("Kind + ComponentTypeSelector combined", func() {
		It("should use Kind for element ID and ComponentTypeSelector for matching", func() {
			workerValue := "worker"
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["role"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			pod := makePod("pod-1", "worker-0", map[string]string{"role": "worker"})
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))

			// ObjectBacked: ID is a UUID, K8sRef has the K8s UID
			Expect(addr.ElementsPath[0].ID).NotTo(BeEmpty())
			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
			Expect(addr.ElementsPath[0].K8sRef.Name).To(Equal("my-sts"))
		})

		It("should skip element when Kind matches but ComponentTypeSelector doesn't", func() {
			workerValue := "worker"
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["role"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			// Pod has "scheduler" role, not "worker"
			pod := makePod("pod-1", "sched-0", map[string]string{"role": "scheduler"})
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			// ComponentTypeSelector takes precedence: pod doesn't match, so no path
			Expect(addr.ElementsPath).To(BeEmpty())
		})
	})

	Context("Kind + ComponentInstanceSelector combined", func() {
		It("should use Kind for element ID and InstanceSelector for instance splitting", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["partition"]`,
						},
					},
				},
			})

			podA := makePod("pod-a", "pod-a", map[string]string{"partition": "shard-1"})
			podB := makePod("pod-b", "pod-b", map[string]string{"partition": "shard-2"})
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addrA := GetPodElementsPath(context.Background(), &podA, karta, ownerChain, nil)
			addrB := GetPodElementsPath(context.Background(), &podB, karta, ownerChain, nil)

			Expect(addrA.ElementsPath).To(HaveLen(1))
			Expect(addrB.ElementsPath).To(HaveLen(1))

			// Both share the same K8sRef UID since Kind is set
			Expect(addrA.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addrB.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addrA.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
			Expect(addrB.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))

			// But instance keys and display names differ
			Expect(addrA.ElementsPath[0].InstanceKey).To(Equal(ptr.To("shard-1")))
			Expect(addrB.ElementsPath[0].InstanceKey).To(Equal(ptr.To("shard-2")))
			Expect(addrA.ElementsPath[0].Name).To(Equal("group:shard-1"))
			Expect(addrB.ElementsPath[0].Name).To(Equal("group:shard-2"))
		})
	})

	Context("Mixed strategies in multi-level hierarchy", func() {
		It("should handle Kind parent with ComponentTypeSelector child", func() {
			workerValue := "worker"
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "StatefulSet",
					},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["role"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			pod := makePod("pod-1", "worker-0", map[string]string{"role": "worker"})
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(2))

			// Level 1: ObjectBacked (Strategy B)
			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
			Expect(addr.ElementsPath[0].K8sRef.Name).To(Equal("my-sts"))

			// Level 2: ComponentTypeSelector (Strategy A) — logical, UUID-based
			Expect(addr.ElementsPath[1].Name).To(Equal("worker"))
			Expect(addr.ElementsPath[1].K8sRef).To(BeNil())
		})

		It("should handle logical parent with Kind child", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "tier", // logical grouping, always matches
				},
				{
					Name:     "replica",
					OwnerRef: ptr.To("tier"),
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "ReplicaSet",
					},
				},
			})

			pod := makePod("pod-1", "pod-0", nil)
			ownerChain := OwnerChain{
				{UID: "rs-uid", Name: "my-rs", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(2))

			// Level 1: logical (Strategy C)
			Expect(addr.ElementsPath[0].Name).To(Equal("tier"))
			Expect(addr.ElementsPath[0].K8sRef).To(BeNil())

			// Level 2: ObjectBacked (Strategy B)
			Expect(addr.ElementsPath[1].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[1].K8sRef.UID).To(Equal(types.UID("rs-uid")))
			Expect(addr.ElementsPath[1].K8sRef.Name).To(Equal("my-rs"))
		})

		It("should handle three-level: Kind → ComponentInstanceSelector → logical", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{
						Group:   "leaderworkerset.x-k8s.io",
						Version: "v1",
						Kind:    "LeaderWorkerSet",
					},
				},
				{
					Name:     "service",
					OwnerRef: ptr.To("group"),
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["service-name"]`,
						},
					},
				},
				{
					Name:     "shard",
					OwnerRef: ptr.To("service"),
				},
			})

			pod := makePod("pod-1", "fe-pod", map[string]string{
				"service-name": "Frontend",
			})
			ownerChain := OwnerChain{
				{UID: "lws-uid", Name: "my-lws", Kind: "LeaderWorkerSet", Group: "leaderworkerset.x-k8s.io", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(3))

			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("lws-uid")))
			Expect(addr.ElementsPath[0].Level).To(Equal(1))

			Expect(addr.ElementsPath[1].InstanceKey).To(Equal(ptr.To("Frontend")))
			Expect(addr.ElementsPath[1].Name).To(Equal("service:frontend"))
			Expect(addr.ElementsPath[1].Level).To(Equal(2))

			Expect(addr.ElementsPath[2].Name).To(Equal("shard"))
			Expect(addr.ElementsPath[2].Level).To(Equal(3))
		})
	})

	Context("ReplicaSelector", func() {
		It("should set ReplicaKey and include replica key in element name", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
			})

			pod := makePod("pod-1", "pod-0", map[string]string{"group-index": "2"})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].Name).To(Equal("group-2"))
			Expect(addr.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("2")))
			Expect(addr.ElementsPath[0].K8sRef).To(BeNil()) // logical element
		})

		It("should produce different elements for different replica keys", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
			})

			pod0 := makePod("pod-0", "pod-0", map[string]string{"group-index": "0"})
			pod1 := makePod("pod-1", "pod-1", map[string]string{"group-index": "1"})

			addr0 := GetPodElementsPath(context.Background(), &pod0, karta, nil, nil)
			addr1 := GetPodElementsPath(context.Background(), &pod1, karta, nil, nil)

			Expect(addr0.ElementsPath).To(HaveLen(1))
			Expect(addr1.ElementsPath).To(HaveLen(1))
			Expect(addr0.ElementsPath[0].Name).To(Equal("group-0"))
			Expect(addr1.ElementsPath[0].Name).To(Equal("group-1"))
			Expect(addr0.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("0")))
			Expect(addr1.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("1")))
			// Different names → different UUIDs
			Expect(addr0.ElementsPath[0].ID).NotTo(Equal(addr1.ElementsPath[0].ID))
		})

		It("should accumulate replica keys across nested levels", func() {
			// Karta: "outer" (ReplicaSelector on group-index) → "inner" (ReplicaSelector on shard-index)
			// Accumulated keys: ["0"] at level 1, ["0", "2"] at level 2
			// Element names: "outer-0" at level 1, "inner-0-2" at level 2
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "outer",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
				{
					Name:     "inner",
					OwnerRef: ptr.To("outer"),
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["shard-index"]`,
						},
					},
				},
			})

			pod := makePod("pod-1", "pod-0", map[string]string{
				"group-index": "0",
				"shard-index": "2",
			})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(2))

			// Level 1: replica key "0"
			Expect(addr.ElementsPath[0].Name).To(Equal("outer-0"))
			Expect(addr.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("0")))
			Expect(addr.ElementsPath[0].Level).To(Equal(1))

			// Level 2: accumulated replica keys "0" + "2" → "inner-0-2"
			Expect(addr.ElementsPath[1].Name).To(Equal("inner-0-2"))
			Expect(addr.ElementsPath[1].ReplicaKey).To(Equal(ptr.To("2")))
			Expect(addr.ElementsPath[1].Level).To(Equal(2))
		})

		It("should set ReplicaKey on Kind-backed element with ReplicaSelector", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["replica-id"]`,
						},
					},
				},
			})

			pod := makePod("pod-1", "worker-0", map[string]string{"replica-id": "3"})
			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}

			addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].Name).To(Equal("worker-3"))
			Expect(addr.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("3")))
			Expect(addr.ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addr.ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
		})

		It("should combine InstanceKey and ReplicaKey in element name", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "service",
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["service-name"]`,
						},
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["replica-id"]`,
						},
					},
				},
			})

			pod := makePod("pod-1", "fe-pod", map[string]string{
				"service-name": "Frontend",
				"replica-id":   "0",
			})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			// generateElementName: "service" + ":Frontend" + "-0" = "service:frontend-0"
			Expect(addr.ElementsPath[0].Name).To(Equal("service:frontend-0"))
			Expect(addr.ElementsPath[0].InstanceKey).To(Equal(ptr.To("Frontend")))
			Expect(addr.ElementsPath[0].ReplicaKey).To(Equal(ptr.To("0")))
		})

		Context("reused parent carries forward ReplicaKey to descendants", func() {
			var karta *v1alpha1.Karta
			var pod corev1.Pod
			var ownerChain OwnerChain
			var existingParent *ewv1alpha1.StructureElement

			BeforeEach(func() {
				karta = makeKarta("root", []v1alpha1.ComponentDefinition{
					{
						Name: "group",
						Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
						PodSelector: &v1alpha1.PodSelector{
							ReplicaSelector: &v1alpha1.ReplicaSelector{
								KeyPath: `.metadata.labels["group-index"]`,
							},
						},
					},
					{
						Name:     "worker",
						OwnerRef: ptr.To("group"),
					},
				})
				pod = makePod("pod-1", "pod-0", map[string]string{"group-index": "0"})
				ownerChain = OwnerChain{
					{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				}
				existingParent = &ewv1alpha1.StructureElement{
					ID:         "existing-group-id",
					Name:       "group-0",
					ReplicaKey: ptr.To("0"),
					K8sRef:     &ewv1alpha1.K8sObjectRef{UID: "sts-uid", Name: "my-sts"},
					Level:      1,
				}
			})

			It("should use accumulated ReplicaKey in new child element name", func() {
				existing := map[string]*ewv1alpha1.StructureElement{
					existingParent.ID: existingParent,
				}
				addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, existing)
				Expect(addr.ElementsPath).To(HaveLen(2))
				Expect(addr.ElementsPath[0].ID).To(Equal("existing-group-id"))
				// Child name includes parent's accumulated replica key
				Expect(addr.ElementsPath[1].Name).To(Equal("worker-0"))
			})

			It("should find existing child by name that includes accumulated ReplicaKey", func() {
				existingChild := &ewv1alpha1.StructureElement{
					ID:       "existing-worker-id",
					Name:     "worker-0",
					ParentID: ptr.To("existing-group-id"),
					Level:    2,
				}
				existing := map[string]*ewv1alpha1.StructureElement{
					existingParent.ID: existingParent,
					existingChild.ID:  existingChild,
				}
				addr := GetPodElementsPath(context.Background(), &pod, karta, ownerChain, existing)
				Expect(addr.ElementsPath).To(HaveLen(2))
				Expect(addr.ElementsPath[0].ID).To(Equal("existing-group-id"))
				// Child is reused (found by name "worker-0"), not created as new "worker"
				Expect(addr.ElementsPath[1].ID).To(Equal("existing-worker-id"))
				Expect(addr.ElementsPath[1].Name).To(Equal("worker-0"))
			})
		})

		It("should not set ReplicaKey when pod lacks the label referenced by ReplicaSelector", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
			})

			// Pod has no "group-index" label
			pod := makePod("pod-1", "pod-0", nil)

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, nil)
			Expect(addr.ElementsPath).To(HaveLen(1))
			Expect(addr.ElementsPath[0].Name).To(Equal("group"))
			Expect(addr.ElementsPath[0].ReplicaKey).To(BeNil())
		})
	})
})

// ============================================================================
// BuildElementStructure Tests
// ============================================================================

var _ = Describe("BuildElementStructure", func() {

	It("should return empty map for empty pods", func() {
		karta := makeKarta("root", nil)
		elements, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, nil, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(elements).To(BeEmpty())
	})

	Context("labeled pods (fast path via existingElementByID)", func() {
		It("should DeepCopy element from existingElementByID", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-uuid-1": {
					ID: "comp-uuid-1", Name: "worker", Level: 1,
				},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "comp-uuid-1",
				}),
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(newAddrs).To(BeEmpty())
			Expect(elements).To(HaveLen(1))
			Expect(elements["comp-uuid-1"]).NotTo(BeNil())
			Expect(elements["comp-uuid-1"].Name).To(Equal("worker"))
			Expect(elements["comp-uuid-1"].Level).To(Equal(1))
		})

		It("should group multiple labeled pods in same element", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-uuid-1": {
					ID: "comp-uuid-1", Name: "worker", Level: 1,
				},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{elementIDLabelKey(1): "comp-uuid-1"}),
				makePod("pod-uid-2", "worker-1", map[string]string{elementIDLabelKey(1): "comp-uuid-1"}),
				makePod("pod-uid-3", "worker-2", map[string]string{elementIDLabelKey(1): "comp-uuid-1"}),
			}

			elements, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(1))
			Expect(elements["comp-uuid-1"]).NotTo(BeNil())
		})

		It("should rebuild multi-level hierarchy from labels", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{Name: "group"},
				{Name: "worker", OwnerRef: ptr.To("group")},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"group-id":  {ID: "group-id", Name: "group", Level: 1},
				"worker-id": {ID: "worker-id", Name: "worker", Level: 2, ParentID: ptr.To("group-id")},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "group-id",
					elementIDLabelKey(2): "worker-id",
				}),
			}

			elements, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(2))

			group := elements["group-id"]
			worker := elements["worker-id"]
			Expect(group.Name).To(Equal("group"))
			Expect(worker.Name).To(Equal("worker"))

			// Parent-child linking rebuilt from labels
			Expect(worker.ParentID).NotTo(BeNil())
			Expect(*worker.ParentID).To(Equal("group-id"))
		})

		It("should rebuild three-level hierarchy from labels", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{Name: "level1"},
				{Name: "level2", OwnerRef: ptr.To("level1")},
				{Name: "level3", OwnerRef: ptr.To("level2")},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"level1-id": {ID: "level1-id", Name: "level1", Level: 1},
				"level2-id": {ID: "level2-id", Name: "level2", Level: 2, ParentID: ptr.To("level1-id")},
				"level3-id": {ID: "level3-id", Name: "level3", Level: 3, ParentID: ptr.To("level2-id")},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "pod-0", map[string]string{
					elementIDLabelKey(1): "level1-id",
					elementIDLabelKey(2): "level2-id",
					elementIDLabelKey(3): "level3-id",
				}),
			}

			elements, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(3))

			Expect(elements["level1-id"].ParentID).To(BeNil())
			Expect(*elements["level2-id"].ParentID).To(Equal("level1-id"))
			Expect(*elements["level3-id"].ParentID).To(Equal("level2-id"))

		})

		It("should DeepCopy full metadata including GVK from existingElementByID", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-uuid-1": {
					ID: "comp-uuid-1", Name: "worker", Level: 1,
					K8sRef: &ewv1alpha1.K8sObjectRef{
						UID:  "sts-k8s-uid",
						Name: "my-sts",
						GVK:  ewv1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					},
				},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "comp-uuid-1",
				}),
			}

			elements, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())

			node := elements["comp-uuid-1"]
			Expect(node.Name).To(Equal("worker"))
			Expect(node.K8sRef).NotTo(BeNil())
			Expect(node.K8sRef.Name).To(Equal("my-sts"))
			Expect(node.K8sRef.GVK.Kind).To(Equal("StatefulSet"))
		})
	})

	Context("unlabeled pods (GetPodElementsPath path)", func() {
		It("should process unlabeled pod via owner chain", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
			})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"worker-0": {
					{UID: "sts-uid-123", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(1))
			Expect(newAddrs).To(HaveLen(1))

			node := findByK8sUID(elements, "sts-uid-123")
			Expect(node).NotTo(BeNil())
			Expect(node.Name).To(Equal("worker"))
			Expect(node.K8sRef).NotTo(BeNil())
			Expect(node.K8sRef.Name).To(Equal("my-sts"))
			Expect(node.K8sRef.GVK.Kind).To(Equal("StatefulSet"))
			Expect(node.Level).To(Equal(1))

			Expect(newAddrs[0].PodID).To(Equal("pod-uid-1"))
			Expect(newAddrs[0].ElementsPath).To(HaveLen(1))
			Expect(newAddrs[0].ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid-123")))
		})

		It("should handle multi-level unlabeled pod via owner chain", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
				{Name: "worker", OwnerRef: ptr.To("group")},
			})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"worker-0": {
					{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(2))
			Expect(newAddrs).To(HaveLen(1))
			Expect(newAddrs[0].ElementsPath).To(HaveLen(2))

			groupID := newAddrs[0].ElementsPath[0].ID
			workerID := newAddrs[0].ElementsPath[1].ID
			Expect(*elements[workerID].ParentID).To(Equal(groupID))
		})

		It("should derive instance key for unlabeled Dynamo pod", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "service",
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["nvidia.com/dynamo-component"]`,
						},
					},
				},
			})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "fe-pod", map[string]string{
					"nvidia.com/dynamo-component": "Frontend",
				}),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"fe-pod": {},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(newAddrs).To(HaveLen(1))

			elemID := newAddrs[0].ElementsPath[0].ID
			Expect(elements[elemID].InstanceKey).To(Equal(ptr.To("Frontend")))
			Expect(elements[elemID].Name).To(Equal("service:frontend"))
		})

		It("should reuse existing element for unlabeled pod when ID matches", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
				{Name: "worker", OwnerRef: ptr.To("group")},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"group-uuid": {
					ID: "group-uuid", Name: "group", Level: 1,
					K8sRef: &ewv1alpha1.K8sObjectRef{
						UID: "sts-uid", Name: "my-sts",
						GVK: ewv1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					},
				},
			}

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"worker-0": {
					{UID: "sts-uid", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(2))
			Expect(newAddrs).To(HaveLen(1))

			// Group reused from existing (DeepCopy)
			group := elements["group-uuid"]
			Expect(group).NotTo(BeNil())
			Expect(group.Name).To(Equal("group"))
			Expect(group.K8sRef.Name).To(Equal("my-sts"))

			// Worker derived via GetPodElementsPath
			workerID := newAddrs[0].ElementsPath[1].ID
			worker := elements[workerID]
			Expect(worker).NotTo(BeNil())
			Expect(worker.Name).To(Equal("worker"))
		})
	})

	Context("mixed labeled and unlabeled pods", func() {
		It("should handle mix of labeled and unlabeled pods in same workload", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-uuid": {
					ID: "comp-uuid", Name: "worker", Level: 1,
					K8sRef: &ewv1alpha1.K8sObjectRef{
						UID: "sts-uid-123", Name: "my-sts",
						GVK: ewv1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					},
				},
			}

			pods := []corev1.Pod{
				// Already labeled (from previous cycle)
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "comp-uuid",
				}),
				// New unlabeled pod (just created)
				makePod("pod-uid-2", "worker-1", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"worker-1": {
					{UID: "sts-uid-123", Name: "my-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
				},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, existing)
			Expect(err).NotTo(HaveOccurred())
			// Both pods resolve to the same element (same StatefulSet UID)
			Expect(elements).To(HaveLen(1))
			Expect(elements["comp-uuid"]).NotTo(BeNil())

			// Only the unlabeled pod gets a newPodAddress
			Expect(newAddrs).To(HaveLen(1))
			Expect(newAddrs[0].PodID).To(Equal("pod-uid-2"))
			Expect(newAddrs[0].ElementsPath[0].ID).To(Equal("comp-uuid"))
		})

		It("should use Dynamo-style instance selector for unlabeled pods", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "service",
					PodSelector: &v1alpha1.PodSelector{
						ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
							IdPath: `.metadata.labels["nvidia.com/dynamo-component"]`,
						},
					},
				},
			})

			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-fe-id": {
					ID: "comp-fe-id", Name: "service:frontend", InstanceKey: ptr.To("Frontend"), Level: 1,
				},
			}

			pods := []corev1.Pod{
				// Already labeled Frontend pod
				makePod("pod-uid-1", "fe-0", map[string]string{
					elementIDLabelKey(1):          "comp-fe-id",
					"nvidia.com/dynamo-component": "Frontend",
				}),
				// New unlabeled PrefillWorker pod
				makePod("pod-uid-2", "pw-0", map[string]string{
					"nvidia.com/dynamo-component": "PrefillWorker",
				}),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pw-0": {},
			}

			elements, newAddrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(2)) // Frontend (labeled) + PrefillWorker (unlabeled)
			Expect(newAddrs).To(HaveLen(1))

			// Frontend reused from existing
			Expect(elements["comp-fe-id"]).NotTo(BeNil())
			Expect(elements["comp-fe-id"].InstanceKey).To(Equal(ptr.To("Frontend")))

			// PrefillWorker created via GetPodElementsPath
			pwAddr := newAddrs[0]
			Expect(pwAddr.ElementsPath).To(HaveLen(1))
			Expect(pwAddr.ElementsPath[0].InstanceKey).To(Equal(ptr.To("PrefillWorker")))
			Expect(pwAddr.ElementsPath[0].Name).To(Equal("service:prefillworker"))

			pwNode := elements[pwAddr.ElementsPath[0].ID]
			Expect(pwNode).NotTo(BeNil())
			Expect(pwNode.InstanceKey).To(Equal(ptr.To("PrefillWorker")))
		})
	})

	Context("input validation", func() {
		It("should reprocess labeled pod with orphaned element IDs when owner chain is provided", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
			})

			// Pod has labels pointing to an element ID that doesn't exist in existing
			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "unknown-comp-id",
				}),
			}

			ownerChain := OwnerChain{
				{UID: "sts-uid", Name: "worker-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
			}
			ownerChainByPodName := map[string]*OwnerChain{"worker-0": &ownerChain}

			elements, addrs, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(elements).To(HaveLen(1))
			Expect(addrs).To(HaveLen(1))

			// The orphaned pod was reprocessed as unlabeled — got a fresh element
			Expect(addrs[0].ElementsPath[0].K8sRef).NotTo(BeNil())
			Expect(addrs[0].ElementsPath[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))
		})

		It("should return error when labeled pod with orphaned IDs has no owner chain", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "unknown-comp-id",
				}),
			}

			_, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("owner chain"))
			Expect(err.Error()).To(ContainSubstring("worker-0"))
		})

		It("should return error when unlabeled pod has no owner chain", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil), // unlabeled, no owner chain
			}

			_, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("owner chain"))
			Expect(err.Error()).To(ContainSubstring("worker-0"))
		})

		It("should return error when unlabeled pod has nil owner chain entry", func() {
			karta := makeKarta("root", []v1alpha1.ComponentDefinition{{Name: "worker"}})

			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"worker-0": nil, // explicitly nil
			}

			_, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no valid entry"))
		})
	})

	Context("PodsRequiringOwnerChain", func() {
		It("should return unlabeled pods", func() {
			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", nil),
				makePod("pod-uid-2", "worker-1", nil),
			}
			result := PodsRequiringOwnerChain(pods, nil)
			Expect(result).To(HaveLen(2))
			Expect(result[0].Name).To(Equal("worker-0"))
			Expect(result[1].Name).To(Equal("worker-1"))
		})

		It("should not return pods with resolvable labels", func() {
			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-1": {ID: "comp-1", Name: "worker"},
			}
			pods := []corev1.Pod{
				makePod("pod-uid-1", "worker-0", map[string]string{
					elementIDLabelKey(1): "comp-1",
				}),
			}
			result := PodsRequiringOwnerChain(pods, existing)
			Expect(result).To(BeEmpty())
		})

		It("should return pods with orphaned element labels", func() {
			existing := map[string]*ewv1alpha1.StructureElement{
				"comp-1": {ID: "comp-1", Name: "worker"},
			}
			pods := []corev1.Pod{
				makePod("pod-uid-1", "labeled-ok", map[string]string{
					elementIDLabelKey(1): "comp-1",
				}),
				makePod("pod-uid-2", "labeled-orphaned", map[string]string{
					elementIDLabelKey(1): "unknown-comp-id",
				}),
				makePod("pod-uid-3", "unlabeled", nil),
			}
			result := PodsRequiringOwnerChain(pods, existing)
			Expect(result).To(HaveLen(2))
			Expect(result[0].Name).To(Equal("labeled-orphaned"))
			Expect(result[1].Name).To(Equal("unlabeled"))
		})

		It("should return nil for empty pods", func() {
			result := PodsRequiringOwnerChain(nil, nil)
			Expect(result).To(BeNil())
		})
	})

	Context("element lifecycle", func() {
		// Two-level Karta with ObjectBacked elements at both levels:
		//   Level 1: "group"  (Kind=StatefulSet)
		//   Level 2: "worker" (Kind=ReplicaSet, OwnerRef="group")
		// Using ObjectBacked at both levels so element IDs come from K8s UIDs,
		// making it easy to verify identity across build cycles.
		var (
			karta *v1alpha1.Karta
			ctx   context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
			karta = makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
					Kind:     &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
				},
			})
		})

		It("should remove element subtrees when their pods are deleted", func() {
			// --- Cycle 1: Build with 5 unlabeled pods across 3 StatefulSets ---
			// Group A: 2 pods, Group B: 2 pods, Group C: 1 pod
			// Expected: 6 elements (3 groups + 3 workers)
			pods := []corev1.Pod{
				makePod("pod-a-0-uid", "pod-a-0", nil),
				makePod("pod-a-1-uid", "pod-a-1", nil),
				makePod("pod-b-0-uid", "pod-b-0", nil),
				makePod("pod-b-1-uid", "pod-b-1", nil),
				makePod("pod-c-0-uid", "pod-c-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-a-1": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-0": {{UID: "rs-B", Name: "rs-b", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-1": {{UID: "rs-B", Name: "rs-b", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-c-0": {{UID: "rs-C", Name: "rs-c", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-C", Name: "sts-c", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle1Elements, cycle1Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle1Elements).To(HaveLen(6))
			Expect(cycle1Addrs).To(HaveLen(5))

			// All 3 groups and 3 workers present
			stsA := findByK8sUID(cycle1Elements, "sts-A")
			stsB := findByK8sUID(cycle1Elements, "sts-B")
			stsC := findByK8sUID(cycle1Elements, "sts-C")
			rsA := findByK8sUID(cycle1Elements, "rs-A")
			rsB := findByK8sUID(cycle1Elements, "rs-B")
			rsC := findByK8sUID(cycle1Elements, "rs-C")
			Expect(stsA).NotTo(BeNil())
			Expect(stsB).NotTo(BeNil())
			Expect(stsC).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())
			Expect(rsB).NotTo(BeNil())
			Expect(rsC).NotTo(BeNil())

			// Label all pods (simulates the caller labeling after Build)
			pods = labelPods(pods, cycle1Addrs)

			// --- Cycle 2: Groups B and C deleted, only A's 2 pods remain ---
			remainingPods := pods[:2] // pod-a-0 and pod-a-1 (now labeled)

			cycle2Elements, cycle2Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, remainingPods, nil, cycle1Elements)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle2Addrs).To(BeEmpty()) // all remaining pods are labeled

			// Only A's subtree remains
			Expect(cycle2Elements).To(HaveLen(2))
			stsA = findByK8sUID(cycle2Elements, "sts-A")
			rsA = findByK8sUID(cycle2Elements, "rs-A")
			Expect(stsA).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())

			// B and C subtrees are gone
			Expect(findByK8sUID(cycle2Elements, "sts-B")).To(BeNil())
			Expect(findByK8sUID(cycle2Elements, "rs-B")).To(BeNil())
			Expect(findByK8sUID(cycle2Elements, "sts-C")).To(BeNil())
			Expect(findByK8sUID(cycle2Elements, "rs-C")).To(BeNil())

			// A's hierarchy preserved with same IDs
			Expect(stsA.Name).To(Equal("group"))
			Expect(rsA.Name).To(Equal("worker"))
			Expect(*rsA.ParentID).To(Equal(stsA.ID))
		})

		It("should assign new element IDs when deleted replicas are recreated with new K8s objects", func() {
			// --- Cycle 1: Build with 4 unlabeled pods across 2 StatefulSets ---
			pods := []corev1.Pod{
				makePod("pod-a-0-uid", "pod-a-0", nil),
				makePod("pod-a-1-uid", "pod-a-1", nil),
				makePod("pod-b-0-uid", "pod-b-0", nil),
				makePod("pod-b-1-uid", "pod-b-1", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-a-1": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-0": {{UID: "rs-B", Name: "rs-b", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-1": {{UID: "rs-B", Name: "rs-b", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle1Elements, cycle1Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle1Elements).To(HaveLen(4)) // 2 groups + 2 workers

			// Remember B's original IDs and names
			stsB := findByK8sUID(cycle1Elements, "sts-B")
			rsB := findByK8sUID(cycle1Elements, "rs-B")
			Expect(stsB).NotTo(BeNil())
			Expect(rsB).NotTo(BeNil())
			originalGroupBName := stsB.Name
			originalWorkerBName := rsB.Name

			pods = labelPods(pods, cycle1Addrs)

			// --- Cycle 2: Delete group B's pods ---
			remainingPods := pods[:2] // pod-a-0, pod-a-1 (labeled)

			cycle2Elements, _, err := BuildElementStructure(ctx, "wl-uid", karta, remainingPods, nil, cycle1Elements)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle2Elements).To(HaveLen(2))
			Expect(findByK8sUID(cycle2Elements, "sts-B")).To(BeNil())
			Expect(findByK8sUID(cycle2Elements, "rs-B")).To(BeNil())

			// --- Cycle 3: Recreate replica with NEW K8s objects (new UIDs) ---
			// The LWS controller creates a new StatefulSet + ReplicaSet with fresh UIDs.
			// The new pods are unlabeled (just created).
			cycle3Pods := []corev1.Pod{
				remainingPods[0],                       // pod-a-0 (labeled)
				remainingPods[1],                       // pod-a-1 (labeled)
				makePod("pod-d-0-uid", "pod-d-0", nil), // new unlabeled
				makePod("pod-d-1-uid", "pod-d-1", nil), // new unlabeled
			}

			newOwnerChainByPodName := map[string]*OwnerChain{
				"pod-d-0": {{UID: "rs-D", Name: "rs-d", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-D", Name: "sts-d", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-d-1": {{UID: "rs-D", Name: "rs-d", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-D", Name: "sts-d", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle3Elements, cycle3Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, cycle3Pods, newOwnerChainByPodName, cycle2Elements)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle3Elements).To(HaveLen(4)) // A's 2 + D's 2
			Expect(cycle3Addrs).To(HaveLen(2))    // only D's pods are new

			// A's elements preserved with same IDs
			stsA := findByK8sUID(cycle3Elements, "sts-A")
			rsA := findByK8sUID(cycle3Elements, "rs-A")
			Expect(stsA).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())

			// D's elements have NEW IDs (different from B's original IDs)
			stsD := findByK8sUID(cycle3Elements, "sts-D")
			rsD := findByK8sUID(cycle3Elements, "rs-D")
			Expect(stsD).NotTo(BeNil())
			Expect(rsD).NotTo(BeNil())
			Expect(stsD.ID).NotTo(Equal(stsB.ID))
			Expect(rsD.ID).NotTo(Equal(rsB.ID))

			// But D's element NAMES are the same as B's — same deterministic name from Karta
			Expect(stsD.Name).To(Equal(originalGroupBName))
			Expect(rsD.Name).To(Equal(originalWorkerBName))

			// Hierarchy correct for recreated elements
			Expect(*rsD.ParentID).To(Equal(stsD.ID))

			// A's hierarchy also still correct
			Expect(*rsA.ParentID).To(Equal(stsA.ID))
		})

		It("fast path: should return existing elements with zero allocation when structure is unchanged", func() {
			// Pre-built structure with hierarchy already linked
			existing := map[string]*ewv1alpha1.StructureElement{
				"uuid-1": {ID: "uuid-1", Name: "group", Level: 1,
					K8sRef: &ewv1alpha1.K8sObjectRef{UID: "sts-A", Name: "sts-a",
						GVK: ewv1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}}},
				"uuid-2": {ID: "uuid-2", Name: "worker", Level: 2,
					ParentID: ptr.To("uuid-1"),
					K8sRef: &ewv1alpha1.K8sObjectRef{UID: "rs-A", Name: "rs-a",
						GVK: ewv1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}}},
			}

			pods := []corev1.Pod{
				makePod("pod-1", "pod-a-0", map[string]string{
					elementIDLabelKey(1): "uuid-1",
					elementIDLabelKey(2): "uuid-2",
				}),
				makePod("pod-2", "pod-a-1", map[string]string{
					elementIDLabelKey(1): "uuid-1",
					elementIDLabelKey(2): "uuid-2",
				}),
			}

			result, newAddrs, err := BuildElementStructure(ctx, "wl-uid", karta, pods, nil, existing)
			Expect(err).NotTo(HaveOccurred())
			Expect(newAddrs).To(BeEmpty())

			// Should return the SAME map and nodes — no DeepCopy
			Expect(result).To(HaveLen(2))
			Expect(result["uuid-1"]).To(BeIdenticalTo(existing["uuid-1"]))
			Expect(result["uuid-2"]).To(BeIdenticalTo(existing["uuid-2"]))

			// Hierarchy untouched
			Expect(*result["uuid-2"].ParentID).To(Equal("uuid-1"))
		})

		It("should scale up: new pod joins existing element without disrupting structure", func() {
			// --- Cycle 1: Build with 2 unlabeled pods under one StatefulSet ---
			pods := []corev1.Pod{
				makePod("pod-a-0-uid", "pod-a-0", nil),
				makePod("pod-a-1-uid", "pod-a-1", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-a-1": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle1Elements, cycle1Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle1Elements).To(HaveLen(2))
			pods = labelPods(pods, cycle1Addrs)

			// --- Cycle 2: Scale up — add a third pod under the SAME StatefulSet ---
			scaledPods := append(pods,
				makePod("pod-a-2-uid", "pod-a-2", nil), // new unlabeled
			)

			scaleUpChains := map[string]*OwnerChain{
				"pod-a-2": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle2Elements, cycle2Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, scaledPods, scaleUpChains, cycle1Elements)
			Expect(err).NotTo(HaveOccurred())

			// Same 2 elements — the new pod joins the existing element, doesn't create new ones
			Expect(cycle2Elements).To(HaveLen(2))
			stsA := findByK8sUID(cycle2Elements, "sts-A")
			rsA := findByK8sUID(cycle2Elements, "rs-A")
			Expect(stsA).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())

			// Only the new pod needs labeling
			Expect(cycle2Addrs).To(HaveLen(1))
			Expect(cycle2Addrs[0].PodID).To(Equal("pod-a-2-uid"))
			Expect(string(cycle2Addrs[0].ElementsPath[0].K8sRef.UID)).To(Equal("sts-A"))
			Expect(string(cycle2Addrs[0].ElementsPath[1].K8sRef.UID)).To(Equal("rs-A"))

			// Hierarchy preserved
			Expect(*rsA.ParentID).To(Equal(stsA.ID))
		})

		It("should preserve element IDs when pods are replaced under the same K8s objects", func() {
			// --- Cycle 1: Build with 2 pods ---
			pods := []corev1.Pod{
				makePod("pod-a-0-uid", "pod-a-0", nil),
				makePod("pod-a-1-uid", "pod-a-1", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-a-1": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle1Elements, cycle1Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle1Elements).To(HaveLen(2))
			pods = labelPods(pods, cycle1Addrs)

			// --- Cycle 2: pod-a-0 killed, pod-a-2 created (new UID, same K8s objects) ---
			// This simulates a pod restart — the StatefulSet and ReplicaSet stay the same.
			churnedPods := []corev1.Pod{
				pods[1],                                // pod-a-1 (labeled, survived)
				makePod("pod-a-2-uid", "pod-a-2", nil), // new unlabeled replacement
			}

			churnChains := map[string]*OwnerChain{
				"pod-a-2": {{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			cycle2Elements, cycle2Addrs, err := BuildElementStructure(ctx, "wl-uid", karta, churnedPods, churnChains, cycle1Elements)
			Expect(err).NotTo(HaveOccurred())

			// Same elements — K8s objects unchanged, so IDs are preserved
			Expect(cycle2Elements).To(HaveLen(2))
			stsA := findByK8sUID(cycle2Elements, "sts-A")
			rsA := findByK8sUID(cycle2Elements, "rs-A")
			Expect(stsA).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())

			// Only the replacement pod needs labeling
			Expect(cycle2Addrs).To(HaveLen(1))
			Expect(cycle2Addrs[0].PodID).To(Equal("pod-a-2-uid"))

			// Hierarchy intact
			Expect(*rsA.ParentID).To(Equal(stsA.ID))
		})

		It("should preserve parent when only one child subtree is removed in a three-level hierarchy", func() {
			// Three-level Karta: lws → group → worker
			threeLevelKarta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "lws",
					Kind: &v1alpha1.GroupVersionKind{Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet"},
				},
				{
					Name:     "group",
					OwnerRef: ptr.To("lws"),
					Kind:     &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
					Kind:     &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
				},
			})

			// Two branches under the same LWS: group-A and group-B
			pods := []corev1.Pod{
				makePod("pod-a-0-uid", "pod-a-0", nil),
				makePod("pod-b-0-uid", "pod-b-0", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {
					{UID: "rs-A", Name: "rs-a", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"},
					{UID: "lws-uid", Name: "my-lws", Kind: "LeaderWorkerSet", Group: "leaderworkerset.x-k8s.io", Version: "v1"},
				},
				"pod-b-0": {
					{UID: "rs-B", Name: "rs-b", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
					{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"},
					{UID: "lws-uid", Name: "my-lws", Kind: "LeaderWorkerSet", Group: "leaderworkerset.x-k8s.io", Version: "v1"},
				},
			}

			cycle1Elements, cycle1Addrs, err := BuildElementStructure(ctx, "wl-uid", threeLevelKarta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cycle1Elements).To(HaveLen(5)) // lws + sts-A + rs-A + sts-B + rs-B

			pods = labelPods(pods, cycle1Addrs)

			// --- Cycle 2: Delete branch B, keep branch A ---
			remainingPods := pods[:1] // pod-a-0 (labeled, 3-level path)

			cycle2Elements, _, err := BuildElementStructure(ctx, "wl-uid", threeLevelKarta, remainingPods, nil, cycle1Elements)
			Expect(err).NotTo(HaveOccurred())

			// LWS parent survives — branch A still references it
			Expect(cycle2Elements).To(HaveLen(3)) // lws + sts-A + rs-A
			lwsNode := findByK8sUID(cycle2Elements, "lws-uid")
			stsA := findByK8sUID(cycle2Elements, "sts-A")
			rsA := findByK8sUID(cycle2Elements, "rs-A")
			Expect(lwsNode).NotTo(BeNil())
			Expect(stsA).NotTo(BeNil())
			Expect(rsA).NotTo(BeNil())

			// Branch B fully removed
			Expect(findByK8sUID(cycle2Elements, "sts-B")).To(BeNil())
			Expect(findByK8sUID(cycle2Elements, "rs-B")).To(BeNil())

			// LWS now only has group-A as child (group-B removed)
			Expect(*stsA.ParentID).To(Equal(lwsNode.ID))
			Expect(*rsA.ParentID).To(Equal(stsA.ID))
		})

		It("should create separate elements for pods with different physical objects of the same Kind", func() {
			// Karta has a single Kind-based component definition (StatefulSet),
			// but 4 pods are owned by 2 different StatefulSets → 2 distinct group elements.
			singleLevelKarta := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
			})

			pods := []corev1.Pod{
				makePod("pod-a-0", "pod-a-0", nil),
				makePod("pod-a-1", "pod-a-1", nil),
				makePod("pod-b-0", "pod-b-0", nil),
				makePod("pod-b-1", "pod-b-1", nil),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "sts-alpha", Name: "sts-alpha", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-a-1": {{UID: "sts-alpha", Name: "sts-alpha", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-0": {{UID: "sts-beta", Name: "sts-beta", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-1": {{UID: "sts-beta", Name: "sts-beta", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			elements, addrs, err := BuildElementStructure(ctx, "wl-uid", singleLevelKarta, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())

			// Two distinct elements from two different StatefulSets
			Expect(elements).To(HaveLen(2))
			stsAlpha := findByK8sUID(elements, "sts-alpha")
			stsBeta := findByK8sUID(elements, "sts-beta")
			Expect(stsAlpha).NotTo(BeNil())
			Expect(stsBeta).NotTo(BeNil())
			Expect(stsAlpha.Name).To(Equal("group"))
			Expect(stsBeta.Name).To(Equal("group"))

			// All 4 pods get addresses
			Expect(addrs).To(HaveLen(4))

			// Label and rebuild — should hit fast path
			pods = labelPods(pods, addrs)
			result, newAddrs, err := BuildElementStructure(ctx, "wl-uid", singleLevelKarta, pods, nil, elements)
			Expect(err).NotTo(HaveOccurred())
			Expect(newAddrs).To(BeEmpty()) // fast path, no new addresses
			Expect(result).To(HaveLen(2))
			Expect(findByK8sUID(result, "sts-alpha")).NotTo(BeNil())
			Expect(findByK8sUID(result, "sts-beta")).NotTo(BeNil())
		})

		It("should resolve unlabeled pods with the same replica key to the same logical parent element", func() {
			// Karta: "group" (logical with ReplicaSelector) → "worker" (Kind=StatefulSet)
			// Two pods owned by different StatefulSets but with the same replica key
			// should share the same logical parent (matched via findElementByName on
			// the deterministic name "group-0").
			multiLevelRI := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
					Kind:     &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
				},
			})

			pods := []corev1.Pod{
				makePod("pod-a-0", "pod-a-0", map[string]string{"group-index": "0"}),
				makePod("pod-b-0", "pod-b-0", map[string]string{"group-index": "0"}),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"pod-a-0": {{UID: "sts-A", Name: "sts-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"pod-b-0": {{UID: "sts-B", Name: "sts-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			elements, addrs, err := BuildElementStructure(ctx, "wl-uid", multiLevelRI, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())

			// 3 elements: 1 shared logical "group-0" + 2 distinct ObjectBacked "worker" elements
			Expect(elements).To(HaveLen(3))
			Expect(addrs).To(HaveLen(2))

			// Both pods share the same logical parent element (same ID, matched by name)
			groupA := addrs[0].ElementsPath[0]
			groupB := addrs[1].ElementsPath[0]
			Expect(groupA.ID).To(Equal(groupB.ID))
			Expect(groupA.Name).To(Equal("group-0"))
			Expect(groupA.ReplicaKey).To(Equal(ptr.To("0")))
			Expect(groupA.K8sRef).To(BeNil()) // logical, not ObjectBacked

			// Worker elements are distinct (different K8s UIDs)
			workerA := addrs[0].ElementsPath[1]
			workerB := addrs[1].ElementsPath[1]
			Expect(workerA.ID).NotTo(Equal(workerB.ID))
			Expect(workerA.K8sRef.UID).To(Equal(types.UID("sts-A")))
			Expect(workerB.K8sRef.UID).To(Equal(types.UID("sts-B")))
			// Workers inherit accumulated replica keys in their name but have no own ReplicaSelector
			Expect(workerA.Name).To(Equal("worker-0"))
			Expect(workerB.Name).To(Equal("worker-0"))
			Expect(workerA.ReplicaKey).To(BeNil())
			Expect(workerB.ReplicaKey).To(BeNil())

			// Both workers have the shared logical group as parent
			Expect(*elements[workerA.ID].ParentID).To(Equal(groupA.ID))
			Expect(*elements[workerB.ID].ParentID).To(Equal(groupA.ID))
		})

		It("should create separate logical elements for pods with different replica keys and preserve ReplicaKey [LWS]", func() {
			// Karta: "group" (logical, ReplicaSelector on group-index) → "worker" (Kind=StatefulSet)
			replicaRI := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "group",
					PodSelector: &v1alpha1.PodSelector{
						ReplicaSelector: &v1alpha1.ReplicaSelector{
							KeyPath: `.metadata.labels["group-index"]`,
						},
					},
				},
				{
					Name:     "leader",
					OwnerRef: ptr.To("group"),
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["worker-index"]`,
							Value:   ptr.To("0"),
						},
					},
				},
				{
					Name:     "worker",
					OwnerRef: ptr.To("group"),
					Kind:     &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["leader-name"]`,
						},
					},
				},
			})

			// 4 pods: 2 in group 0 (leader+worker), 2 in group 1 (leader+worker)
			pods := []corev1.Pod{
				makePod("leader-a-0", "leader-a-0", map[string]string{"group-index": "0", "worker-index": "0"}),
				makePod("worker-a-0", "worker-a-0", map[string]string{"group-index": "0", "worker-index": "1",
					"leader-name": "leader-0"}),
				makePod("leader-b-0", "leader-b-0", map[string]string{"group-index": "1", "worker-index": "0"}),
				makePod("worker-b-0", "worker-b-0", map[string]string{"group-index": "1", "worker-index": "1",
					"leader-name": "leader-1"}),
			}

			ownerChainByPodName := map[string]*OwnerChain{
				"leader-a-0": {{UID: "sts-leaders", Name: "sts-leaders", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"worker-a-0": {{UID: "sts-workers-A", Name: "sts-workers-a", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"leader-b-0": {{UID: "sts-leaders", Name: "sts-leaders", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"worker-b-0": {{UID: "sts-workers-B", Name: "sts-workers-b", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			elements, addrs, err := BuildElementStructure(ctx, "lws-uid", replicaRI, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())

			// 6 elements: group-0, group-1 (logical), leader-a, worker-a, leader-b, worker-b
			Expect(elements).To(HaveLen(6))
			Expect(addrs).To(HaveLen(4))

			// Find the two group and two leader elements by name
			var leader0, leader1 *ewv1alpha1.StructureElement
			var group0, group1 *ewv1alpha1.StructureElement
			for _, e := range elements {
				if e.Name == "group-0" {
					group0 = e
				}
				if e.Name == "group-1" {
					group1 = e
				}
				if e.Name == "leader-0" {
					leader0 = e
				}
				if e.Name == "leader-1" {
					leader1 = e
				}
			}
			Expect(group0).NotTo(BeNil())
			Expect(group1).NotTo(BeNil())
			Expect(group0.ID).NotTo(Equal(group1.ID))
			Expect(group0.ReplicaKey).To(Equal(ptr.To("0")))
			Expect(group1.ReplicaKey).To(Equal(ptr.To("1")))
			Expect(group0.K8sRef).To(BeNil()) // logical
			Expect(group1.K8sRef).To(BeNil()) // logical
			Expect(leader0).NotTo(BeNil())
			Expect(leader1).NotTo(BeNil())
			Expect(leader0.ID).NotTo(Equal(leader1.ID))
			Expect(leader0.ReplicaKey).To(BeNil())
			Expect(leader1.ReplicaKey).To(BeNil())
			Expect(leader0.K8sRef).To(BeNil()) // logical
			Expect(leader1.K8sRef).To(BeNil()) // logical
			Expect(leader0.ParentID).To(Equal(ptr.To(group0.ID)))
			Expect(leader1.ParentID).To(Equal(ptr.To(group1.ID)))

			// Worker elements are ObjectBacked
			stsWorkerA := findByK8sUID(elements, "sts-workers-A")
			stsWorkerB := findByK8sUID(elements, "sts-workers-B")
			Expect(stsWorkerA).NotTo(BeNil())
			Expect(stsWorkerB).NotTo(BeNil())
			Expect(stsWorkerA.Name).To(Equal("worker-0"))
			Expect(stsWorkerB.Name).To(Equal("worker-1"))

			// Workers have the correct group as parent
			Expect(*stsWorkerA.ParentID).To(Equal(group0.ID))
			Expect(*stsWorkerB.ParentID).To(Equal(group1.ID))

			// Leader are logical - not object backed, so no K8sRef, but should still have correct parent

			// Label and rebuild — should hit fast path for labeled pods
			pods = labelPods(pods, addrs)
			result, newAddrs, err := BuildElementStructure(ctx, "wl-uid", replicaRI, pods, nil, elements)
			Expect(err).NotTo(HaveOccurred())
			Expect(newAddrs).To(BeEmpty())
			Expect(result).To(HaveLen(6))
		})

		It("should handle Kind + ComponentTypeSelector distinguishing multiple pod roles under different owners", func() {
			// Two component definitions with the same Kind (StatefulSet) but different
			// ComponentTypeSelectors — leader vs worker. Each is backed by its own StatefulSet.
			leaderValue := "leader"
			workerValue := "worker"
			mixedRI := makeKarta("root", []v1alpha1.ComponentDefinition{
				{
					Name: "leader",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["role"]`,
							Value:   &leaderValue,
						},
					},
				},
				{
					Name: "worker",
					Kind: &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
					PodSelector: &v1alpha1.PodSelector{
						ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
							KeyPath: `.metadata.labels["role"]`,
							Value:   &workerValue,
						},
					},
				},
			})

			pods := []corev1.Pod{
				makePod("leader-pod-0", "leader-pod-0", map[string]string{"role": "leader"}),
				makePod("worker-pod-0", "worker-pod-0", map[string]string{"role": "worker"}),
				makePod("worker-pod-1", "worker-pod-1", map[string]string{"role": "worker"}),
			}

			// Leader and worker pods are under different StatefulSets
			ownerChainByPodName := map[string]*OwnerChain{
				"leader-pod-0": {{UID: "sts-leader", Name: "sts-leader", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"worker-pod-0": {{UID: "sts-worker", Name: "sts-worker", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
				"worker-pod-1": {{UID: "sts-worker", Name: "sts-worker", Kind: "StatefulSet", Group: "apps", Version: "v1"}},
			}

			elements, addrs, err := BuildElementStructure(ctx, "wl-uid", mixedRI, pods, ownerChainByPodName, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(elements).To(HaveLen(2))
			stsLeader := findByK8sUID(elements, "sts-leader")
			stsWorker := findByK8sUID(elements, "sts-worker")
			Expect(stsLeader).NotTo(BeNil())
			Expect(stsWorker).NotTo(BeNil())
			Expect(stsLeader.Name).To(Equal("leader"))
			Expect(stsWorker.Name).To(Equal("worker"))
			Expect(addrs).To(HaveLen(3))
		})
	})
})

// ============================================================================
// GeneratePodElementLabels Tests
// ============================================================================

var _ = Describe("GeneratePodElementLabels", func() {

	It("should generate single -id label for single level", func() {
		path := []ewv1alpha1.StructureElement{
			{
				ID:    "comp-1",
				Name:  "worker",
				Level: 1,
			},
		}

		labels := GeneratePodElementLabels(path)
		Expect(labels).To(HaveLen(1))
		Expect(labels[elementIDLabelKey(1)]).To(Equal("comp-1"))
	})

	It("should generate -id labels for multiple levels", func() {
		path := []ewv1alpha1.StructureElement{
			{
				ID:    "group-id",
				Name:  "group",
				Level: 1,
			},
			{
				ID:          "service-fe-id",
				Name:        "service:frontend",
				InstanceKey: ptr.To("Frontend"),
				Level:       2,
			},
		}

		labels := GeneratePodElementLabels(path)
		Expect(labels).To(HaveLen(2))
		Expect(labels[elementIDLabelKey(1)]).To(Equal("group-id"))
		Expect(labels[elementIDLabelKey(2)]).To(Equal("service-fe-id"))
	})

	It("should skip levels outside 1-3", func() {
		path := []ewv1alpha1.StructureElement{
			{ID: "root", Level: 0},
			{ID: "comp-1", Level: 1},
			{ID: "deep", Level: 4},
		}

		labels := GeneratePodElementLabels(path)
		Expect(labels).To(HaveLen(1))
		Expect(labels[elementIDLabelKey(1)]).To(Equal("comp-1"))
	})
})

// ============================================================================
// IsStructureUnchanged Tests
// ============================================================================

var _ = Describe("IsStructureUnchanged", func() {

	It("should return false when existingElementsByID is empty", func() {
		pods := []corev1.Pod{makePod("pod-1", "my-pod", nil)}
		Expect(IsStructureUnchanged(pods, nil)).To(BeFalse())
		Expect(IsStructureUnchanged(pods, map[string]*ewv1alpha1.StructureElement{})).To(BeFalse())
	})

	It("should return false when pods is empty", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"comp-1": {ID: "comp-1", Name: "worker"},
		}
		Expect(IsStructureUnchanged(nil, existing)).To(BeFalse())
		Expect(IsStructureUnchanged([]corev1.Pod{}, existing)).To(BeFalse())
	})

	It("should return false when any pod is unlabeled", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"comp-1": {ID: "comp-1", Name: "worker"},
		}
		pods := []corev1.Pod{
			makePod("pod-1", "labeled-pod", map[string]string{elementIDLabelKey(1): "comp-1"}),
			makePod("pod-2", "unlabeled-pod", nil),
		}
		Expect(IsStructureUnchanged(pods, existing)).To(BeFalse())
	})

	It("should return false when labeled pod references unknown element", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"comp-1": {ID: "comp-1", Name: "worker"},
		}
		pods := []corev1.Pod{
			makePod("pod-1", "my-pod", map[string]string{elementIDLabelKey(1): "unknown-id"}),
		}
		Expect(IsStructureUnchanged(pods, existing)).To(BeFalse())
	})

	It("should return false when pods reference a subset of existing elements", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"comp-1": {ID: "comp-1", Name: "group"},
			"comp-2": {ID: "comp-2", Name: "worker"},
		}
		pods := []corev1.Pod{
			makePod("pod-1", "my-pod", map[string]string{elementIDLabelKey(1): "comp-1"}),
		}
		Expect(IsStructureUnchanged(pods, existing)).To(BeFalse())
	})

	It("should return true when all pods labeled and reference exactly the existing elements", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"comp-1": {ID: "comp-1", Name: "group"},
			"comp-2": {ID: "comp-2", Name: "worker"},
		}
		pods := []corev1.Pod{
			makePod("pod-1", "pod-a", map[string]string{
				elementIDLabelKey(1): "comp-1",
				elementIDLabelKey(2): "comp-2",
			}),
			makePod("pod-2", "pod-b", map[string]string{
				elementIDLabelKey(1): "comp-1",
				elementIDLabelKey(2): "comp-2",
			}),
		}
		Expect(IsStructureUnchanged(pods, existing)).To(BeTrue())
	})

	It("should return true when no pods and no existing elements", func() {
		Expect(IsStructureUnchanged(nil, nil)).To(BeTrue())
		Expect(IsStructureUnchanged([]corev1.Pod{}, map[string]*ewv1alpha1.StructureElement{})).To(BeTrue())
	})
})

// ============================================================================
// ValidateKartaDepth Tests
// ============================================================================

var _ = Describe("ValidateKartaDepth", func() {

	It("should accept nil Karta", func() {
		Expect(ValidateKartaDepth(nil)).NotTo(HaveOccurred())
	})

	It("should accept empty ChildComponents", func() {
		karta := makeKarta("root", nil)
		Expect(ValidateKartaDepth(karta)).NotTo(HaveOccurred())
	})

	It("should accept depth 1", func() {
		karta := makeKarta("root", []v1alpha1.ComponentDefinition{
			{Name: "child1"},
		})
		Expect(ValidateKartaDepth(karta)).NotTo(HaveOccurred())
	})

	It("should accept depth 2", func() {
		karta := makeKarta("root", []v1alpha1.ComponentDefinition{
			{Name: "child1"},
			{Name: "child2", OwnerRef: ptr.To("child1")},
		})
		Expect(ValidateKartaDepth(karta)).NotTo(HaveOccurred())
	})

	It("should accept depth 3", func() {
		karta := makeKarta("root", []v1alpha1.ComponentDefinition{
			{Name: "child1"},
			{Name: "child2", OwnerRef: ptr.To("child1")},
			{Name: "child3", OwnerRef: ptr.To("child2")},
		})
		Expect(ValidateKartaDepth(karta)).NotTo(HaveOccurred())
	})

	It("should reject depth 4", func() {
		karta := makeKarta("root", []v1alpha1.ComponentDefinition{
			{Name: "child1"},
			{Name: "child2", OwnerRef: ptr.To("child1")},
			{Name: "child3", OwnerRef: ptr.To("child2")},
			{Name: "child4", OwnerRef: ptr.To("child3")},
		})
		err := ValidateKartaDepth(karta)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("nesting depth 4"))
		Expect(err.Error()).To(ContainSubstring("exceeds maximum"))
	})
})

// ============================================================================
// Helper Function Tests
// ============================================================================

var _ = Describe("Helper Functions", func() {

	Context("elementIDLabelKey", func() {
		It("should build correct label key", func() {
			Expect(elementIDLabelKey(1)).To(Equal("run.ai/element-1-id"))
			Expect(elementIDLabelKey(2)).To(Equal("run.ai/element-2-id"))
			Expect(elementIDLabelKey(3)).To(Equal("run.ai/element-3-id"))
		})
	})

	Context("generateElementName", func() {
		It("should generate name with Karta name only", func() {
			Expect(generateElementName("worker", nil, nil)).To(Equal("worker"))
		})

		It("should generate name with instance", func() {
			Expect(generateElementName("service", ptr.To("Frontend"), nil)).To(Equal("service:frontend"))
		})

		It("should generate name with replica", func() {
			Expect(generateElementName("group", nil, []string{"0"})).To(Equal("group-0"))
		})

		It("should generate name with instance and replica", func() {
			Expect(generateElementName("service", ptr.To("Frontend"), []string{"0"})).To(Equal("service:frontend-0"))
		})

		It("should generate name with nested replicas", func() {
			Expect(generateElementName("worker", nil, []string{"0", "2"})).To(Equal("worker-0-2"))
		})
	})

	Context("findPhysicalObjectFromChain", func() {
		It("should find matching Kind in chain", func() {
			chain := OwnerChain{
				{UID: "rs-uid", Name: "my-rs", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
				{UID: "deploy-uid", Name: "my-deploy", Kind: "Deployment", Group: "apps", Version: "v1"},
			}
			target := &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			ref := findPhysicalObjectFromChain(target, chain)
			Expect(ref).NotTo(BeNil())
			Expect(ref.UID).To(Equal(types.UID("deploy-uid")))
			Expect(ref.Name).To(Equal("my-deploy"))
		})

		It("should return nil when Kind not in chain", func() {
			chain := OwnerChain{
				{UID: "rs-uid", Name: "my-rs", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
			}
			target := &v1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}
			ref := findPhysicalObjectFromChain(target, chain)
			Expect(ref).To(BeNil())
		})

		It("should return nil for nil target GVK", func() {
			chain := OwnerChain{{UID: "uid", Name: "foo", Kind: "Foo", Version: "v1"}}
			ref := findPhysicalObjectFromChain(nil, chain)
			Expect(ref).To(BeNil())
		})

		It("should match core API group correctly", func() {
			chain := OwnerChain{
				{UID: "svc-uid", Name: "my-svc", Kind: "Service", Version: "v1"},
			}
			target := &v1alpha1.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}
			ref := findPhysicalObjectFromChain(target, chain)
			Expect(ref).NotTo(BeNil())
			Expect(ref.UID).To(Equal(types.UID("svc-uid")))
			Expect(ref.Name).To(Equal("my-svc"))
		})
	})

	Context("podMatchesComponentTypeSelector", func() {
		It("should match when label value matches", func() {
			expectedValue := "worker"
			selector := &v1alpha1.PodSelector{
				ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
					KeyPath: `.metadata.labels["role"]`,
					Value:   &expectedValue,
				},
			}
			pod := makePod("pod-1", "p", map[string]string{"role": "worker"})
			Expect(podMatchesComponentTypeSelector(context.Background(), &pod, selector)).To(BeTrue())
		})

		It("should not match when label value differs", func() {
			expectedValue := "worker"
			selector := &v1alpha1.PodSelector{
				ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
					KeyPath: `.metadata.labels["role"]`,
					Value:   &expectedValue,
				},
			}
			pod := makePod("pod-1", "p", map[string]string{"role": "scheduler"})
			Expect(podMatchesComponentTypeSelector(context.Background(), &pod, selector)).To(BeFalse())
		})

		It("should match when key exists and no value specified", func() {
			selector := &v1alpha1.PodSelector{
				ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
					KeyPath: `.metadata.labels["role"]`,
				},
			}
			pod := makePod("pod-1", "p", map[string]string{"role": "anything"})
			Expect(podMatchesComponentTypeSelector(context.Background(), &pod, selector)).To(BeTrue())
		})

		It("should not match when key doesn't exist", func() {
			selector := &v1alpha1.PodSelector{
				ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
					KeyPath: `.metadata.labels["role"]`,
				},
			}
			pod := makePod("pod-1", "p", map[string]string{"other": "value"})
			Expect(podMatchesComponentTypeSelector(context.Background(), &pod, selector)).To(BeFalse())
		})

		It("should return false for nil selector", func() {
			pod := makePod("pod-1", "p", nil)
			Expect(podMatchesComponentTypeSelector(context.Background(), &pod, nil)).To(BeFalse())
		})
	})

	Context("extractInstanceKey", func() {
		It("should extract instance key from label", func() {
			pod := makePod("pod-1", "p", map[string]string{"nvidia.com/dynamo-component": "Frontend"})
			selector := &v1alpha1.PodSelector{
				ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
					IdPath: `.metadata.labels["nvidia.com/dynamo-component"]`,
				},
			}
			Expect(extractInstanceKey(context.Background(), &pod, selector)).To(Equal(ptr.To("Frontend")))
		})

		It("should extract instance key from annotation", func() {
			pod := makePodWithAnnotations("pod-1", "p", nil, map[string]string{"my-annot": "my-instance"})
			selector := &v1alpha1.PodSelector{
				ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
					IdPath: `.metadata.annotations["my-annot"]`,
				},
			}
			Expect(extractInstanceKey(context.Background(), &pod, selector)).To(Equal(ptr.To("my-instance")))
		})

		It("should return nil for missing label", func() {
			pod := makePod("pod-1", "p", map[string]string{"other": "value"})
			selector := &v1alpha1.PodSelector{
				ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
					IdPath: `.metadata.labels["missing"]`,
				},
			}
			Expect(extractInstanceKey(context.Background(), &pod, selector)).To(BeNil())
		})

		It("should return nil for nil selector", func() {
			pod := makePod("pod-1", "p", nil)
			Expect(extractInstanceKey(context.Background(), &pod, nil)).To(BeNil())
		})

		It("should return nil when no ComponentInstanceSelector", func() {
			pod := makePod("pod-1", "p", nil)
			selector := &v1alpha1.PodSelector{}
			Expect(extractInstanceKey(context.Background(), &pod, selector)).To(BeNil())
		})
	})

	Context("topologicalSortComponents", func() {
		It("should sort independent components in stable order", func() {
			components := []v1alpha1.ComponentDefinition{
				{Name: "a"},
				{Name: "b"},
			}
			sorted := topologicalSortComponents(components)
			Expect(sorted).To(HaveLen(2))
			names := []string{sorted[0].Name, sorted[1].Name}
			Expect(names).To(ContainElements("a", "b"))
		})

		It("should sort parent before child", func() {
			components := []v1alpha1.ComponentDefinition{
				{Name: "child", OwnerRef: ptr.To("parent")},
				{Name: "parent"},
			}
			sorted := topologicalSortComponents(components)
			Expect(sorted).To(HaveLen(2))
			Expect(sorted[0].Name).To(Equal("parent"))
			Expect(sorted[1].Name).To(Equal("child"))
		})

		It("should handle deep chain", func() {
			components := []v1alpha1.ComponentDefinition{
				{Name: "c", OwnerRef: ptr.To("b")},
				{Name: "a"},
				{Name: "b", OwnerRef: ptr.To("a")},
			}
			sorted := topologicalSortComponents(components)
			Expect(sorted).To(HaveLen(3))
			Expect(sorted[0].Name).To(Equal("a"))
			Expect(sorted[1].Name).To(Equal("b"))
			Expect(sorted[2].Name).To(Equal("c"))
		})

		It("should treat component whose OwnerRef points to root (not in ChildComponents) as root-level", func() {
			components := []v1alpha1.ComponentDefinition{
				{Name: "replicaset", OwnerRef: ptr.To("deployment")},
			}
			sorted := topologicalSortComponents(components)
			Expect(sorted).To(HaveLen(1))
			Expect(sorted[0].Name).To(Equal("replicaset"))
		})

		It("should handle mix of root-parented and child-parented components", func() {
			// "replicaset" OwnerRef points to root "deployment" (not in ChildComponents)
			// "container" OwnerRef points to "replicaset" (IS in ChildComponents)
			components := []v1alpha1.ComponentDefinition{
				{Name: "container", OwnerRef: ptr.To("replicaset")},
				{Name: "replicaset", OwnerRef: ptr.To("deployment")},
			}
			sorted := topologicalSortComponents(components)
			Expect(sorted).To(HaveLen(2))
			Expect(sorted[0].Name).To(Equal("replicaset"))
			Expect(sorted[1].Name).To(Equal("container"))
		})

		It("should return nil for empty input", func() {
			Expect(topologicalSortComponents(nil)).To(BeNil())
		})
	})

	Context("maxChildDepth", func() {
		It("should return 0 for no children", func() {
			children := map[string][]string{}
			Expect(maxChildDepth("root", children, 0)).To(Equal(0))
		})

		It("should return 1 for single child", func() {
			children := map[string][]string{
				"root": {"child1"},
			}
			Expect(maxChildDepth("root", children, 0)).To(Equal(1))
		})

		It("should return max depth across branches", func() {
			children := map[string][]string{
				"root": {"a", "b"},
				"a":    {"a1"},
				"a1":   {"a1x"},
				"b":    {"b1"},
			}
			Expect(maxChildDepth("root", children, 0)).To(Equal(3))
		})
	})
})

// ============================================================================
// Dynamo End-to-End Test
// ============================================================================

var _ = Describe("Dynamo End-to-End", func() {
	It("should correctly handle Dynamo-style Karta with multiple service instances", func() {
		karta := makeKarta("dynamo-graph", []v1alpha1.ComponentDefinition{
			{
				Name: "service",
				PodSelector: &v1alpha1.PodSelector{
					ComponentInstanceSelector: &v1alpha1.ComponentInstanceSelector{
						IdPath: `.metadata.labels["nvidia.com/dynamo-component"]`,
					},
				},
			},
		})

		pods := []struct {
			uid       string
			name      string
			component string
		}{
			{"pod-fe-0", "frontend-0", "Frontend"},
			{"pod-fe-1", "frontend-1", "Frontend"},
			{"pod-pw-0", "prefill-0", "PrefillWorker"},
			{"pod-pw-1", "prefill-1", "PrefillWorker"},
			{"pod-dw-0", "decode-0", "DecodeWorker"},
		}

		instanceToElementID := map[string]string{}
		allPodAddresses := []*PodAddress{}
		existing := map[string]*ewv1alpha1.StructureElement{}

		for _, p := range pods {
			pod := makePod(p.uid, p.name, map[string]string{
				"nvidia.com/dynamo-component": p.component,
			})

			addr := GetPodElementsPath(context.Background(), &pod, karta, nil, existing)
			Expect(addr.ElementsPath).To(HaveLen(1))
			allPodAddresses = append(allPodAddresses, addr)

			element := addr.ElementsPath[0]
			if _, found := instanceToElementID[p.component]; !found {
				instanceToElementID[p.component] = element.ID
				existing[element.ID] = &ewv1alpha1.StructureElement{
					ID:          element.ID,
					Name:        element.Name,
					InstanceKey: element.InstanceKey,
				}
			}
		}

		// Verify 3 distinct element instances
		Expect(instanceToElementID).To(HaveLen(3))
		Expect(instanceToElementID).To(HaveKey("Frontend"))
		Expect(instanceToElementID).To(HaveKey("PrefillWorker"))
		Expect(instanceToElementID).To(HaveKey("DecodeWorker"))

		// Verify different element IDs per instance
		feID := instanceToElementID["Frontend"]
		pwID := instanceToElementID["PrefillWorker"]
		dwID := instanceToElementID["DecodeWorker"]
		Expect(feID).NotTo(Equal(pwID))
		Expect(feID).NotTo(Equal(dwID))
		Expect(pwID).NotTo(Equal(dwID))

		// Verify all pods have correct instance keys
		for _, addr := range allPodAddresses {
			Expect(addr.ElementsPath[0].InstanceKey).NotTo(BeNil())
		}
	})
})
