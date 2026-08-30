package structure

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	ewv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconsts "github.com/run-ai/runai/runai-cluster/common/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// ============================================================================
// Test Infrastructure
// ============================================================================

func updaterTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(ewv1alpha1.AddToScheme(s))
	utilruntime.Must(kartav1alpha1.AddToScheme(s))
	return s
}

func newUpdaterTestClient(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&ewv1alpha1.ExternalWorkload{}).
		Build()
}

// makeEW creates an ExternalWorkload with the given name, namespace, uid, and originalUID.
func makeEW(name, ns, uid, originalUID string) *ewv1alpha1.ExternalWorkload {
	return &ewv1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uid),
		},
		Spec: ewv1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID: types.UID(originalUID),
		},
	}
}

// makeWorkloadPod creates a pod with the cluster-top-owner-uid label and optional ownerRefs.
func makeWorkloadPod(uid, name, ns, ewOriginalUID string, extraLabels map[string]string, ownerRefs []metav1.OwnerReference) *corev1.Pod {
	labels := map[string]string{
		commonconsts.ClusterTopOwnerUidLabel: ewOriginalUID,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:             types.UID(uid),
			Name:            name,
			Namespace:       ns,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
	}
}

// makeKartaForGVK creates a Karta with GVK labels and RootComponent.Kind set.
func makeKartaForGVK(rootName string, gvk schema.GroupVersionKind, children []kartav1alpha1.ComponentDefinition) *kartav1alpha1.Karta {
	karta := &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name: rootName + "-karta",
			Labels: map[string]string{
				ewv1alpha1.KartaGroupLabel:   gvk.Group,
				ewv1alpha1.KartaVersionLabel: gvk.Version,
				ewv1alpha1.KartaKindLabel:    gvk.Kind,
			},
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: rootName,
					Kind: &kartav1alpha1.GroupVersionKind{
						Group:   gvk.Group,
						Version: gvk.Version,
						Kind:    gvk.Kind,
					},
				},
				ChildComponents: children,
			},
		},
	}
	return karta
}

// makeOwnerRef creates an OwnerReference for use in pod/owner chain setup.
func makeOwnerRef(name, uid, kind, apiVersion string) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		UID:        types.UID(uid),
	}
}

// ============================================================================
// existingElementsToMap Tests
// ============================================================================

var _ = Describe("existingElementsToMap", func() {
	It("should return nil for nil input", func() {
		result := existingElementsToMap(nil)
		Expect(result).To(BeNil())
	})

	It("should return empty map for empty elements", func() {
		result := existingElementsToMap(&ewv1alpha1.StructureElements{
			Elements: []ewv1alpha1.StructureElement{},
		})
		Expect(result).To(BeEmpty())
	})

	It("should convert multiple elements to map keyed by ID", func() {
		se := &ewv1alpha1.StructureElements{
			Elements: []ewv1alpha1.StructureElement{
				{ID: "aaa", Name: "worker", Level: 1},
				{ID: "bbb", Name: "shard", Level: 2},
			},
		}
		result := existingElementsToMap(se)
		Expect(result).To(HaveLen(2))
		Expect(result["aaa"]).NotTo(BeNil())
		Expect(result["aaa"].Name).To(Equal("worker"))
		Expect(result["bbb"]).NotTo(BeNil())
		Expect(result["bbb"].Name).To(Equal("shard"))
	})
})

// ============================================================================
// elementsMapToSortedSlice Tests
// ============================================================================

var _ = Describe("elementsMapToSortedSlice", func() {
	It("should return empty slice for empty map", func() {
		result := elementsMapToSortedSlice(map[string]*ewv1alpha1.StructureElement{})
		Expect(result).To(BeEmpty())
	})

	It("should sort by level ascending, then ID ascending", func() {
		m := map[string]*ewv1alpha1.StructureElement{
			"zzz": {ID: "zzz", Level: 1},
			"aaa": {ID: "aaa", Level: 2},
			"mmm": {ID: "mmm", Level: 1},
			"bbb": {ID: "bbb", Level: 2},
		}
		result := elementsMapToSortedSlice(m)
		Expect(result).To(HaveLen(4))
		Expect(result[0].ID).To(Equal("mmm"))
		Expect(result[1].ID).To(Equal("zzz"))
		Expect(result[2].ID).To(Equal("aaa"))
		Expect(result[3].ID).To(Equal("bbb"))
	})
})

// ============================================================================
// sortElements Tests
// ============================================================================

var _ = Describe("sortElements", func() {
	It("should not panic on nil slice", func() {
		Expect(func() { sortElements(nil) }).NotTo(Panic())
	})

	It("should not panic on empty slice", func() {
		Expect(func() { sortElements([]ewv1alpha1.StructureElement{}) }).NotTo(Panic())
	})

	It("should sort by level ascending, then ID ascending", func() {
		elements := []ewv1alpha1.StructureElement{
			{ID: "c", Level: 2},
			{ID: "a", Level: 1},
			{ID: "b", Level: 2},
			{ID: "d", Level: 1},
		}
		sortElements(elements)
		Expect(elements[0].ID).To(Equal("a"))
		Expect(elements[1].ID).To(Equal("d"))
		Expect(elements[2].ID).To(Equal("b"))
		Expect(elements[3].ID).To(Equal("c"))
	})
})

// ============================================================================
// structureElementsPatchBytes Tests
// ============================================================================

var _ = Describe("structureElementsPatchBytes", func() {
	It("should produce valid JSON with status.structureElements wrapper", func() {
		se := &ewv1alpha1.StructureElements{
			Elements: []ewv1alpha1.StructureElement{
				{ID: "elem-1", Name: "worker", Level: 1},
			},
			LastCalculated: metav1.Now(),
		}

		patchBytes, err := structureElementsPatchBytes(se)
		Expect(err).NotTo(HaveOccurred())

		var parsed map[string]interface{}
		Expect(json.Unmarshal(patchBytes, &parsed)).To(Succeed())

		status, ok := parsed["status"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		_, ok = status["structureElements"]
		Expect(ok).To(BeTrue())
	})
})

// ============================================================================
// buildOwnerChain Tests
// ============================================================================

var _ = Describe("buildOwnerChain", func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
	)

	BeforeEach(func() {
		scheme = updaterTestScheme()
		ctx = context.Background()
	})

	It("should return empty chain for pod with no OwnerRefs", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lonely-pod",
				Namespace: "ns",
				UID:       "pod-uid",
			},
		}
		c := newUpdaterTestClient(scheme, pod)
		updater := &StructureUpdater{cachedClient: c}

		chain, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(chain).To(BeEmpty())
	})

	It("should return single-entry chain for pod -> ReplicaSet (1 level)", func() {
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		c := newUpdaterTestClient(scheme, pod, rs)
		updater := &StructureUpdater{cachedClient: c}

		chain, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(chain).To(HaveLen(1))
		Expect(chain[0].UID).To(Equal(types.UID("rs-uid")))
		Expect(chain[0].Name).To(Equal("my-rs"))
		Expect(chain[0].Kind).To(Equal("ReplicaSet"))
		Expect(chain[0].Group).To(Equal("apps"))
		Expect(chain[0].Version).To(Equal("v1"))
	})

	It("should walk multi-level chain: pod -> ReplicaSet -> StatefulSet", func() {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sts",
				Namespace: "ns",
				UID:       "sts-uid",
			},
		}
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-sts", "sts-uid", "StatefulSet", "apps/v1"),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		c := newUpdaterTestClient(scheme, pod, rs, sts)
		updater := &StructureUpdater{cachedClient: c}

		chain, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(chain).To(HaveLen(2))

		Expect(chain[0].Kind).To(Equal("ReplicaSet"))
		Expect(chain[0].UID).To(Equal(types.UID("rs-uid")))

		Expect(chain[1].Kind).To(Equal("StatefulSet"))
		Expect(chain[1].UID).To(Equal(types.UID("sts-uid")))
	})

	It("should return partial chain when owner is forbidden", func() {
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-runner", "runner-uid", "WorkloadRunner", "run.ai/v1"),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		c := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(pod, rs).
			WithStatusSubresource(&ewv1alpha1.ExternalWorkload{}).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key.Name == "my-runner" {
						return apierrors.NewForbidden(
							schema.GroupResource{Group: "run.ai", Resource: "workloadrunners"},
							"my-runner",
							fmt.Errorf("User cannot get resource"),
						)
					}
					return client.Get(ctx, key, obj, opts...)
				},
			}).
			Build()
		updater := &StructureUpdater{cachedClient: c}

		chain, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(chain).To(HaveLen(1))
		Expect(chain[0].Kind).To(Equal("ReplicaSet"))
		Expect(chain[0].UID).To(Equal(types.UID("rs-uid")))
	})

	It("should return error when owner not found", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("missing-rs", "missing-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		c := newUpdaterTestClient(scheme, pod)
		updater := &StructureUpdater{cachedClient: c}

		_, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed getting owner"))
	})

	It("should return error when owner has multiple OwnerRefs", func() {
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("owner-a", "uid-a", "StatefulSet", "apps/v1"),
					makeOwnerRef("owner-b", "uid-b", "Deployment", "apps/v1"),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		c := newUpdaterTestClient(scheme, pod, rs)
		updater := &StructureUpdater{cachedClient: c}

		_, err := updater.buildOwnerChain(ctx, pod)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("multiple owners"))
	})
})

// ============================================================================
// fetchMissingOwnerChains Tests
// ============================================================================

var _ = Describe("fetchMissingOwnerChains", func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
	)

	BeforeEach(func() {
		scheme = updaterTestScheme()
		ctx = context.Background()
	})

	It("should return nil when all pods have resolvable element labels", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"elem-1": {ID: "elem-1", Name: "worker", Level: 1},
		}

		pods := []corev1.Pod{
			makePod("pod-1", "pod-a", map[string]string{
				elementIDLabelKey(1): "elem-1",
			}),
		}

		c := newUpdaterTestClient(scheme)
		updater := &StructureUpdater{cachedClient: c}

		chains, err := updater.fetchMissingOwnerChains(ctx, pods, existing)
		Expect(err).NotTo(HaveOccurred())
		Expect(chains).To(BeNil())
	})

	It("should fetch chains only for unlabeled pods", func() {
		existing := map[string]*ewv1alpha1.StructureElement{
			"elem-1": {ID: "elem-1", Name: "worker", Level: 1},
		}

		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
			},
		}

		labeledPod := makePod("pod-labeled", "labeled-pod", map[string]string{
			elementIDLabelKey(1): "elem-1",
		})
		unlabeledPod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "pod-unlabeled",
				Name:      "unlabeled-pod",
				Namespace: "ns",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}

		pods := []corev1.Pod{labeledPod, unlabeledPod}

		c := newUpdaterTestClient(scheme, rs)
		updater := &StructureUpdater{cachedClient: c}

		chains, err := updater.fetchMissingOwnerChains(ctx, pods, existing)
		Expect(err).NotTo(HaveOccurred())
		Expect(chains).To(HaveLen(1))
		Expect(chains).To(HaveKey("unlabeled-pod"))
		Expect(*chains["unlabeled-pod"]).To(HaveLen(1))
		Expect((*chains["unlabeled-pod"])[0].Kind).To(Equal("ReplicaSet"))
	})

	It("should propagate error when owner chain fetch fails", func() {
		unlabeledPod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "pod-unlabeled",
				Name:      "unlabeled-pod",
				Namespace: "ns",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("missing-rs", "missing-uid", "ReplicaSet", "apps/v1"),
				},
			},
		}
		pods := []corev1.Pod{unlabeledPod}

		c := newUpdaterTestClient(scheme)
		updater := &StructureUpdater{cachedClient: c}

		_, err := updater.fetchMissingOwnerChains(ctx, pods, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed building owner chain"))
	})
})

// ============================================================================
// patchStructureStatus Tests
// ============================================================================

var _ = Describe("patchStructureStatus", func() {
	It("should patch EW status with sorted elements and lastCalculated", func() {
		scheme := updaterTestScheme()
		ctx := context.Background()

		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")
		c := newUpdaterTestClient(scheme, ew)
		updater := &StructureUpdater{cachedClient: c}

		elementsByID := map[string]*ewv1alpha1.StructureElement{
			"bbb": {ID: "bbb", Name: "shard", Level: 2},
			"aaa": {ID: "aaa", Name: "worker", Level: 1},
		}

		err := updater.patchStructureStatus(ctx, ew, elementsByID)
		Expect(err).NotTo(HaveOccurred())

		// Re-fetch to verify
		fetched := &ewv1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "my-ew", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.StructureElements).NotTo(BeNil())
		Expect(fetched.Status.StructureElements.Elements).To(HaveLen(2))
		// Sorted: level 1 first, then level 2
		Expect(fetched.Status.StructureElements.Elements[0].ID).To(Equal("aaa"))
		Expect(fetched.Status.StructureElements.Elements[1].ID).To(Equal("bbb"))
		Expect(fetched.Status.StructureElements.LastCalculated.IsZero()).To(BeFalse())
	})
})

// ============================================================================
// labelPodsWithElements Tests
// ============================================================================

var _ = Describe("labelPodsWithElements", func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
	)

	BeforeEach(func() {
		scheme = updaterTestScheme()
		ctx = context.Background()
	})

	It("should be no-op for nil podAddresses", func() {
		c := newUpdaterTestClient(scheme)
		updater := &StructureUpdater{cachedClient: c}

		// Should not panic
		updater.labelPodsWithElements(ctx, nil, nil)
	})

	It("should be no-op for empty podAddresses", func() {
		c := newUpdaterTestClient(scheme)
		updater := &StructureUpdater{cachedClient: c}

		updater.labelPodsWithElements(ctx, []corev1.Pod{}, []*PodAddress{})
	})

	It("should label pods with element labels", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid-1",
			},
		}

		c := newUpdaterTestClient(scheme, pod)
		updater := &StructureUpdater{cachedClient: c}

		pods := []corev1.Pod{*pod}
		addresses := []*PodAddress{
			{
				PodID: "pod-uid-1",
				ElementsPath: []ewv1alpha1.StructureElement{
					{ID: "elem-aaa", Level: 1},
				},
			},
		}

		updater.labelPodsWithElements(ctx, pods, addresses)

		// Re-fetch pod
		fetched := &corev1.Pod{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "my-pod", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Labels).To(HaveKeyWithValue(elementIDLabelKey(1), "elem-aaa"))
	})

	It("should skip PodAddress with UID not in pod list", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "ns",
				UID:       "pod-uid-1",
			},
		}

		c := newUpdaterTestClient(scheme, pod)
		updater := &StructureUpdater{cachedClient: c}

		pods := []corev1.Pod{*pod}
		addresses := []*PodAddress{
			{
				PodID: "nonexistent-uid",
				ElementsPath: []ewv1alpha1.StructureElement{
					{ID: "elem-aaa", Level: 1},
				},
			},
		}

		// Should not panic
		updater.labelPodsWithElements(ctx, pods, addresses)

		// Pod should remain unlabeled
		fetched := &corev1.Pod{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "my-pod", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Labels).To(BeEmpty())
	})
})

// ============================================================================
// RecalculateStructure Integration Tests
// ============================================================================

var _ = Describe("RecalculateStructure", func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		stsGVK schema.GroupVersionKind
	)

	BeforeEach(func() {
		scheme = updaterTestScheme()
		ctx = context.Background()
		stsGVK = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}
	})

	It("should return nil when EW not found", func() {
		c := newUpdaterTestClient(scheme)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "missing-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return nil when EW exists but no pods", func() {
		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")
		c := newUpdaterTestClient(scheme, ew)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "my-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return nil when structure is unchanged (all pods labeled)", func() {
		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")
		ew.Status.StructureElements = &ewv1alpha1.StructureElements{
			Elements: []ewv1alpha1.StructureElement{
				{ID: "elem-1", Name: "worker", Level: 1},
			},
		}

		pod := makeWorkloadPod("pod-1", "my-pod", "ns", "top-uid",
			map[string]string{elementIDLabelKey(1): "elem-1"}, nil)

		c := newUpdaterTestClient(scheme, ew, pod)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "my-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle happy path: unlabeled pod -> structure built, EW patched, pod labeled", func() {
		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")

		karta := makeKartaForGVK("root", stsGVK, []kartav1alpha1.ComponentDefinition{
			{
				Name: "worker",
				Kind: &kartav1alpha1.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "StatefulSet",
				},
			},
		})

		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sts",
				Namespace: "ns",
				UID:       "sts-uid",
			},
		}
		rs := &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-rs",
				Namespace: "ns",
				UID:       "rs-uid",
				OwnerReferences: []metav1.OwnerReference{
					makeOwnerRef("my-sts", "sts-uid", "StatefulSet", "apps/v1"),
				},
			},
		}

		pod := makeWorkloadPod("pod-uid-1", "my-pod", "ns", "top-uid",
			nil,
			[]metav1.OwnerReference{
				makeOwnerRef("my-rs", "rs-uid", "ReplicaSet", "apps/v1"),
			})

		c := newUpdaterTestClient(scheme, ew, karta, sts, rs, pod)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "my-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).NotTo(HaveOccurred())

		// Verify EW status has structure elements
		fetchedEW := &ewv1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "my-ew", Namespace: "ns"}, fetchedEW)).To(Succeed())
		Expect(fetchedEW.Status.StructureElements).NotTo(BeNil())
		Expect(fetchedEW.Status.StructureElements.Elements).To(HaveLen(1))
		Expect(fetchedEW.Status.StructureElements.Elements[0].Name).To(Equal("worker"))
		Expect(fetchedEW.Status.StructureElements.Elements[0].K8sRef).NotTo(BeNil())
		Expect(fetchedEW.Status.StructureElements.Elements[0].K8sRef.UID).To(Equal(types.UID("sts-uid")))

		// Verify pod has element label
		fetchedPod := &corev1.Pod{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "my-pod", Namespace: "ns"}, fetchedPod)).To(Succeed())
		Expect(fetchedPod.Labels).To(HaveKey(elementIDLabelKey(1)))
	})

	It("should return nil when Karta not found", func() {
		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")

		// Pod with no element labels — will be treated as unlabeled, but no ownerRefs needed
		// because the code path is: fetch pods -> unchanged check (false) -> fetch owner chains ->
		// buildStructureAndPatchStatus -> getRI (not found) -> return nil.
		// The pod needs no ownerRefs here because Karta-not-found short-circuits before BuildElementStructure.
		pod := makeWorkloadPod("pod-uid-1", "my-pod", "ns", "top-uid", nil, nil)

		c := newUpdaterTestClient(scheme, ew, pod)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "my-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return error when owner chain fails (owner missing)", func() {
		ew := makeEW("my-ew", "ns", "ew-uid", "top-uid")

		pod := makeWorkloadPod("pod-uid-1", "my-pod", "ns", "top-uid",
			nil,
			[]metav1.OwnerReference{
				makeOwnerRef("missing-rs", "missing-uid", "ReplicaSet", "apps/v1"),
			})

		c := newUpdaterTestClient(scheme, ew, pod)
		updater := &StructureUpdater{cachedClient: c}

		err := updater.RecalculateStructure(ctx,
			types.NamespacedName{Name: "my-ew", Namespace: "ns"},
			stsGVK)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed building owner chain"))
	})
})
