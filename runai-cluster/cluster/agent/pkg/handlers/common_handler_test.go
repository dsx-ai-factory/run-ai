/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"sync/atomic"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CommonHandler", func() {
	var (
		origin   client.Object
		existing client.Object
	)

	BeforeEach(func() {
		// Create a basic Pod object for testing
		origin = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
			},
		}
		existing = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
			},
		}
	})
	Context("existingPreCopy", func() {
		It("exist annotation is nil - should not modify origin annotations", func() {
			existing.SetAnnotations(nil)
			origin.SetAnnotations(map[string]string{"existing-key": "existing-value"})

			existingPreCopy(origin, existing)

			Expect(origin.GetAnnotations()).To(Equal(map[string]string{"existing-key": "existing-value"}))
		})

		It("exist annotation is empty should not modify origin annotations", func() {
			existing.SetAnnotations(map[string]string{})
			origin.SetAnnotations(map[string]string{"existing-key": "existing-value"})

			existingPreCopy(origin, existing)

			Expect(origin.GetAnnotations()).To(Equal(map[string]string{"existing-key": "existing-value"}))
		})

		It("exist annotation is not topology one - should not modify origin annotations", func() {
			existing.SetAnnotations(map[string]string{
				"other-annotation": "other-value",
				"another-key":      "another-value",
			})
			origin.SetAnnotations(map[string]string{"existing-key": "existing-value"})

			existingPreCopy(origin, existing)

			Expect(origin.GetAnnotations()).To(Equal(map[string]string{"existing-key": "existing-value"}))
		})

		Context("copy existing topology annotations", func() {
			It("exist has topology annotations - should copy topology name annotation to origin", func() {
				existing.SetAnnotations(map[string]string{
					topology.KAITopologyAnnotationKey: "topology-name",
				})
				origin.SetAnnotations(nil)

				existingPreCopy(origin, existing)

				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					topology.KAITopologyAnnotationKey: "topology-name",
				}))
			})

			It("exist has topology preferred annotations - should copy topology preferred annotation to origin", func() {
				existing.SetAnnotations(map[string]string{
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
				})
				origin.SetAnnotations(nil)

				existingPreCopy(origin, existing)

				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
				}))
			})

			It("exist has topology required annotations - should copy topology required annotation to origin", func() {
				existing.SetAnnotations(map[string]string{
					topology.KAITopologyAnnotationRequiredKey: "required-placement",
				})
				origin.SetAnnotations(nil)

				existingPreCopy(origin, existing)

				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					topology.KAITopologyAnnotationRequiredKey: "required-placement",
				}))
			})

			It("exist has all topology annotations - should copy all topology annotations to origin", func() {
				existing.SetAnnotations(map[string]string{
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
					topology.KAITopologyAnnotationRequiredKey:  "required-placement",
				})
				origin.SetAnnotations(nil)

				existingPreCopy(origin, existing)

				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
					topology.KAITopologyAnnotationRequiredKey:  "required-placement",
				}))
			})

			It("exist has topology annotations and other annotations - should copy only topology annotations to origin", func() {
				existing.SetAnnotations(map[string]string{
					"other-annotation":                         "other-value",
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
					"another-key":                              "another-value",
				})
				origin.SetAnnotations(nil)

				existingPreCopy(origin, existing)

				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
				}))
			})

			It("exist has topology annotations and origin other annotations - should preserve existing annotations and add topology annotations", func() {
				// Arrange
				existing.SetAnnotations(map[string]string{
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
				})
				origin.SetAnnotations(map[string]string{
					"existing-key": "existing-value",
					"another-key":  "another-value",
				})

				// Act
				existingPreCopy(origin, existing)

				// Assert
				Expect(origin.GetAnnotations()).To(Equal(map[string]string{
					"existing-key":                             "existing-value",
					"another-key":                              "another-value",
					topology.KAITopologyAnnotationKey:          "topology-name",
					topology.KAITopologyAnnotationPreferredKey: "preferred-placement",
				}))
			})
		})
	})

	Context("WorkloadRunner restart state preservation on sync", func() {
		const (
			requestRefValue = "rr-uid-1234"
			managedByValue  = "restart-controller-workloadrunner"
		)

		var inspector UnstructuredObjectInspector

		BeforeEach(func() {
			inspector = UnstructuredObjectInspector{}
		})

		// newWorkloadRunner builds a minimal unstructured WorkloadRunner with the given labels and spec.
		newWorkloadRunner := func(labels, spec map[string]interface{}) *unstructured.Unstructured {
			metadata := map[string]interface{}{
				"name":      "freeze-2",
				"namespace": "runai-teama",
			}
			if labels != nil {
				metadata["labels"] = labels
			}
			return &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": ApiVersionWorkloadRunner,
				"kind":       KindWorkloadRunner,
				"metadata":   metadata,
				"spec":       spec,
			}}
		}

		// nodeAffinityInstruction is the accumulated in-cluster node-exclusion instruction,
		// as stored under spec.instructions.nodeAffinity.
		nodeAffinityInstruction := func() map[string]interface{} {
			return map[string]interface{}{
				"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
					"nodeSelectorTerms": []interface{}{
						map[string]interface{}{
							"matchExpressions": []interface{}{
								map[string]interface{}{
									"key":      "kubernetes.io/hostname",
									"operator": "NotIn",
									"values":   []interface{}{"node-a"},
								},
							},
						},
					},
				},
			}
		}

		It("preserves request-ref, managedBy and nodeAffinity mid-restart while keeping CP-composed siblings", func() {
			existingWLR := newWorkloadRunner(
				map[string]interface{}{
					requestRefLabel:                   requestRefValue,
					"run.ai/existing-in-cluster-only": "should-not-survive", // not in the allowlist
				},
				map[string]interface{}{
					"managedBy": managedByValue,
					"instructions": map[string]interface{}{
						"nodeAffinity": nodeAffinityInstruction(),
					},
				},
			)
			// CP payload: no restart state, but carries its own label and an mnnvl instruction.
			originWLR := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{
					"instructions": map[string]interface{}{
						"gpuOptimizations": map[string]interface{}{
							"mnnvl": map[string]interface{}{"mode": "Required"},
						},
					},
				},
			)

			inspector.DeepCopySpec(originWLR, existingWLR)

			// existingWLR is the object that would be Update()d.
			// request-ref preserved, CP label kept, non-allowlisted in-cluster label NOT kept.
			Expect(existingWLR.GetLabels()).To(HaveKeyWithValue(requestRefLabel, requestRefValue))
			Expect(existingWLR.GetLabels()).To(HaveKeyWithValue("run.ai/managed-by-cp", "true"))
			Expect(existingWLR.GetLabels()).NotTo(HaveKey("run.ai/existing-in-cluster-only"))

			// managedBy preserved.
			managedBy, found, err := unstructured.NestedString(existingWLR.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(managedBy).To(Equal(managedByValue))

			// nodeAffinity preserved.
			na, found, err := unstructured.NestedMap(existingWLR.Object, "spec", "instructions", "nodeAffinity")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(na).To(Equal(nodeAffinityInstruction()))

			// CP-composed sibling instruction (mnnvl) applied on the final object.
			mnnvlMode, found, err := unstructured.NestedString(existingWLR.Object, "spec", "instructions", "gpuOptimizations", "mnnvl", "mode")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(mnnvlMode).To(Equal("Required"))
		})

		It("keeps the payload pristine and re-applies nothing after handback", func() {
			existingWLR := newWorkloadRunner(
				map[string]interface{}{},
				map[string]interface{}{"managedBy": ""}, // cleared by the handoff reconciler at handback
			)
			originWLR := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{
					"instructions": map[string]interface{}{
						"gpuOptimizations": map[string]interface{}{
							"mnnvl": map[string]interface{}{"mode": "Required"},
						},
					},
				},
			)

			inspector.DeepCopySpec(originWLR, existingWLR)

			Expect(existingWLR.GetLabels()).NotTo(HaveKey(requestRefLabel))
			// managedBy is not injected (empty in-cluster value does not win over CP's unset).
			_, found, err := unstructured.NestedString(existingWLR.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
			// nodeAffinity is not injected.
			_, found, err = unstructured.NestedFieldNoCopy(existingWLR.Object, "spec", "instructions", "nodeAffinity")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("preserves restart state across a suspend/resume CP write", func() {
			existingWLR := newWorkloadRunner(
				map[string]interface{}{requestRefLabel: requestRefValue},
				map[string]interface{}{
					"managedBy":    managedByValue,
					"instructions": map[string]interface{}{"nodeAffinity": nodeAffinityInstruction()},
					"template":     map[string]interface{}{"active": true},
				},
			)
			// CP payload represents a resume/suspend (active flips) - the one legitimate CP spec write for v1-cast.
			originWLR := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{"template": map[string]interface{}{"active": false}},
			)

			inspector.DeepCopySpec(originWLR, existingWLR)

			Expect(existingWLR.GetLabels()).To(HaveKeyWithValue(requestRefLabel, requestRefValue))
			managedBy, _, _ := unstructured.NestedString(existingWLR.Object, "spec", "managedBy")
			Expect(managedBy).To(Equal(managedByValue))
			na, found, _ := unstructured.NestedMap(existingWLR.Object, "spec", "instructions", "nodeAffinity")
			Expect(found).To(BeTrue())
			Expect(na).To(Equal(nodeAffinityInstruction()))
			// CP's suspend/resume intent (active=false) is applied.
			active, _, _ := unstructured.NestedBool(existingWLR.Object, "spec", "template", "active")
			Expect(active).To(BeFalse())
		})

		It("does not preserve restart state on non-WorkloadRunner objects", func() {
			existingOther := &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": ApiVersionWorkloadRunner,
				"kind":       "SomethingElse",
				"metadata": map[string]interface{}{
					"name":      "x",
					"namespace": "y",
					"labels":    map[string]interface{}{requestRefLabel: requestRefValue},
				},
				"spec": map[string]interface{}{"managedBy": managedByValue},
			}}
			originOther := &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": ApiVersionWorkloadRunner,
				"kind":       "SomethingElse",
				"metadata":   map[string]interface{}{"name": "x", "namespace": "y"},
				"spec":       map[string]interface{}{},
			}}

			inspector.DeepCopySpec(originOther, existingOther)

			Expect(existingOther.GetLabels()).NotTo(HaveKey(requestRefLabel))
			_, found, err := unstructured.NestedString(existingOther.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("does not leak restart state from a conflicted attempt into a retry against a clean live object", func() {
			// The CP payload is built once per sync and reused across RetryOnConflict
			// attempts, so DeepCopySpec must never mutate it.
			payload := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{"template": map[string]interface{}{"active": true}},
			)

			// Attempt 1: the live object is mid-restart; the write then hits a conflict.
			midRestart := newWorkloadRunner(
				map[string]interface{}{requestRefLabel: requestRefValue},
				map[string]interface{}{
					"managedBy":    managedByValue,
					"instructions": map[string]interface{}{"nodeAffinity": nodeAffinityInstruction()},
				},
			)
			inspector.DeepCopySpec(payload, midRestart)

			// The payload stays pristine between attempts.
			Expect(payload.GetLabels()).NotTo(HaveKey(requestRefLabel))
			_, found, err := unstructured.NestedString(payload.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
			_, found, err = unstructured.NestedFieldNoCopy(payload.Object, "spec", "instructions")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())

			// Attempt 2: the refetched live object is clean - handback happened between attempts.
			handedBack := newWorkloadRunner(
				map[string]interface{}{},
				map[string]interface{}{"managedBy": ""},
			)
			inspector.DeepCopySpec(payload, handedBack)

			// The handed-back WLR must not be re-frozen with attempt 1's stale state.
			Expect(handedBack.GetLabels()).NotTo(HaveKey(requestRefLabel))
			_, found, err = unstructured.NestedString(handedBack.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
			_, found, err = unstructured.NestedFieldNoCopy(handedBack.Object, "spec", "instructions", "nodeAffinity")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("does not clobber a CP-provided managedBy when the in-cluster value is empty", func() {
			existingWLR := newWorkloadRunner(
				map[string]interface{}{},
				map[string]interface{}{"managedBy": ""},
			)
			originWLR := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{"managedBy": "cp-managed-by"},
			)

			inspector.DeepCopySpec(originWLR, existingWLR)

			managedBy, found, err := unstructured.NestedString(existingWLR.Object, "spec", "managedBy")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(managedBy).To(Equal("cp-managed-by"))
		})

		It("creates spec.instructions when preserving nodeAffinity over a payload without instructions", func() {
			existingWLR := newWorkloadRunner(
				map[string]interface{}{},
				map[string]interface{}{"instructions": map[string]interface{}{"nodeAffinity": nodeAffinityInstruction()}},
			)
			originWLR := newWorkloadRunner(
				map[string]interface{}{"run.ai/managed-by-cp": "true"},
				map[string]interface{}{},
			)

			inspector.DeepCopySpec(originWLR, existingWLR)

			na, found, err := unstructured.NestedMap(existingWLR.Object, "spec", "instructions", "nodeAffinity")
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(na).To(Equal(nodeAffinityInstruction()))
		})
	})

	Context("getObjectByGVK with retry on conflict", func() {
		var (
			mockClient      *mockConflictClient
			handler         *CommonHandler
			testPod         *corev1.Pod
			testGVK         schema.GroupVersionKind
			testUnstruct    *unstructured.Unstructured
			unstructGVK     schema.GroupVersionKind
			unstructClient  *mockConflictClient
			unstructHandler *CommonHandler
		)

		BeforeEach(func() {
			// Set up for structured (typed) object tests
			testGVK = schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			}
			testPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-pod",
					Namespace:       "test-namespace",
					ResourceVersion: "1",
				},
			}
			testPod.SetGroupVersionKind(testGVK)

			mockClient = newMockConflictClient(testPod)
			handler = &CommonHandler{
				runaiClient:    mockClient,
				cachedClient:   mockClient, // Use mockClient for both since we're testing locally
				gvk:            testGVK,
				isUnstructured: false,
			}

			// Set up for unstructured object tests
			unstructGVK = schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			}
			testUnstruct = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":            "test-deployment",
						"namespace":       "test-namespace",
						"resourceVersion": "1",
					},
				},
			}
			testUnstruct.SetGroupVersionKind(unstructGVK)

			unstructClient = newMockConflictClient(testUnstruct)
			unstructHandler = &CommonHandler{
				runaiClient:    unstructClient,
				cachedClient:   unstructClient, // Use unstructClient for both since we're testing locally
				isUnstructured: true,
			}
		})

		It("should successfully fetch object for structured handler", func() {
			obj, err := handler.getObjectByGVK("test-pod", "test-namespace", testGVK)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj).ToNot(BeNil())
			Expect(obj.GetName()).To(Equal("test-pod"))
			Expect(obj.GetNamespace()).To(Equal("test-namespace"))
		})

		It("should successfully fetch object for unstructured handler", func() {
			obj, err := unstructHandler.getObjectByGVK("test-deployment", "test-namespace", unstructGVK)
			Expect(err).ToNot(HaveOccurred())
			Expect(obj).ToNot(BeNil())
			Expect(obj.GetName()).To(Equal("test-deployment"))
			Expect(obj.GetNamespace()).To(Equal("test-namespace"))

			// Verify it's an unstructured object
			u, ok := obj.(*unstructured.Unstructured)
			Expect(ok).To(BeTrue())
			Expect(u.GetKind()).To(Equal("Deployment"))
		})

		It("should retry and refetch on conflict for structured handler", func() {
			// Configure mock to return conflict on first update, then succeed
			mockClient.conflictsBeforeSuccess = 2

			// Simulate handleExisting update logic
			existingObj := testPod.DeepCopy()
			existingObj.ResourceVersion = "1"

			updateAttempts := 0
			err := mockClient.Update(context.Background(), existingObj)

			for errors.IsConflict(err) && updateAttempts < 5 {
				// Refetch latest version
				latestObj, refetchErr := handler.getObjectByGVK("test-pod", "test-namespace", testGVK)
				Expect(refetchErr).ToNot(HaveOccurred())

				// The mock increments resource version on each get (starting from 1, adds 1 each call)
				// So after N calls, the resource version should be N+1
				expectedVersion := fmt.Sprintf("%d", atomic.LoadInt32(&mockClient.getCallCount)+1)
				Expect(latestObj.GetResourceVersion()).To(Equal(expectedVersion))

				// Try update again with latest object
				updateAttempts++
				err = mockClient.Update(context.Background(), latestObj)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(updateAttempts).To(Equal(2), "Should have retried 2 times before success")
			Expect(mockClient.getCallCount).To(BeNumerically(">=", 2), "Should have called Get at least twice")
		})

		It("should retry and refetch on conflict for unstructured handler", func() {
			// Configure mock to return conflict on first update, then succeed
			unstructClient.conflictsBeforeSuccess = 2

			// Simulate handleExisting update logic
			existingObj := testUnstruct.DeepCopy()
			existingObj.SetResourceVersion("1")

			updateAttempts := 0
			err := unstructClient.Update(context.Background(), existingObj)

			for errors.IsConflict(err) && updateAttempts < 5 {
				// Refetch latest version
				latestObj, refetchErr := unstructHandler.getObjectByGVK("test-deployment", "test-namespace", unstructGVK)
				Expect(refetchErr).ToNot(HaveOccurred())

				// The mock increments resource version on each get (starting from 1, adds 1 each call)
				// So after N calls, the resource version should be N+1
				expectedVersion := fmt.Sprintf("%d", atomic.LoadInt32(&unstructClient.getCallCount)+1)
				Expect(latestObj.GetResourceVersion()).To(Equal(expectedVersion))

				// Try update again with latest object
				updateAttempts++
				err = unstructClient.Update(context.Background(), latestObj)
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(updateAttempts).To(Equal(2), "Should have retried 2 times before success")
			Expect(unstructClient.getCallCount).To(BeNumerically(">=", 2), "Should have called Get at least twice")
		})

		It("should use cached client for fetch", func() {
			// The getObjectByGVK uses h.cachedClient.Get (which is sufficient since we refetch immediately before update)
			// This is verified by the mock client being called
			obj, err := handler.getObjectByGVK("test-pod", "test-namespace", testGVK)
			Expect(err).ToNot(HaveOccurred())
			Expect(mockClient.getCallCount).To(Equal(int32(1)))
			Expect(obj).ToNot(BeNil())
		})
	})
})

// mockConflictClient simulates a Kubernetes client that returns conflict errors on updates
type mockConflictClient struct {
	object                 client.Object
	conflictsBeforeSuccess int32
	updateCallCount        int32
	getCallCount           int32
	currentResourceVersion int32
}

func newMockConflictClient(obj client.Object) *mockConflictClient {
	return &mockConflictClient{
		object:                 obj,
		conflictsBeforeSuccess: 0,
		currentResourceVersion: 1,
	}
}

func (m *mockConflictClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	atomic.AddInt32(&m.getCallCount, 1)

	// Increment resource version on each get to simulate concurrent modifications
	newVersion := atomic.AddInt32(&m.currentResourceVersion, 1)

	// Copy the stored object to the provided object
	if pod, ok := obj.(*corev1.Pod); ok {
		srcPod := m.object.(*corev1.Pod)
		*pod = *srcPod.DeepCopy()
		pod.ResourceVersion = fmt.Sprintf("%d", newVersion)
		return nil
	}

	if u, ok := obj.(*unstructured.Unstructured); ok {
		srcU := m.object.(*unstructured.Unstructured)
		*u = *srcU.DeepCopy()
		u.SetResourceVersion(fmt.Sprintf("%d", newVersion))
		return nil
	}

	return fmt.Errorf("unsupported object type in mock")
}

func (m *mockConflictClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	callCount := atomic.AddInt32(&m.updateCallCount, 1)

	// Return conflict error for the first N calls
	if callCount <= m.conflictsBeforeSuccess {
		return errors.NewConflict(
			schema.GroupResource{Group: "", Resource: "pods"},
			obj.GetName(),
			fmt.Errorf("the object has been modified; please apply your changes to the latest version and try again"),
		)
	}

	// Success after configured number of conflicts
	return nil
}

func (m *mockConflictClient) Scheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	return scheme
}

// Implement other client.Client methods as no-ops for this test
func (m *mockConflictClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

func (m *mockConflictClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return nil
}

func (m *mockConflictClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return nil
}

func (m *mockConflictClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (m *mockConflictClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (m *mockConflictClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return nil
}

func (m *mockConflictClient) Status() client.StatusWriter {
	return nil
}

func (m *mockConflictClient) SubResource(subResource string) client.SubResourceClient {
	return nil
}

func (m *mockConflictClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (m *mockConflictClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return true, nil
}

func (m *mockConflictClient) RESTMapper() meta.RESTMapper {
	return nil
}
