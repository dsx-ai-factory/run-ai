package cachedclient

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCachedClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CachedClient Suite")
}

// Minimal fake implementing our local CachedInformer interface only (not controller-runtime's Informer)
type testCachedInformer struct{ synced bool }

func (f *testCachedInformer) AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error) {
	return nil, nil
}
func (f *testCachedInformer) HasSynced() bool { return f.synced }

func makeGVK(group, version, kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: group, Version: version, Kind: kind}
}

func makeUnstructured(gvk schema.GroupVersionKind) *unstructuredObj {
	u := &unstructuredObj{}
	u.SetGroupVersionKind(gvk)
	return u
}

// unstructuredObj is a minimal object implementing client.Object with GVK support for testing
type unstructuredObj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (u *unstructuredObj) DeepCopyObject() runtime.Object { c := *u; return &c }

// ---------------------------
// Tests
// ---------------------------

var _ = Describe("ExternalWorkloadsCachedClient", func() {
	var (
		ewc    *ExternalWorkloadsCachedClient
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())
		writer := fake.NewClientBuilder().WithScheme(scheme).Build()
		ewc = &ExternalWorkloadsCachedClient{
			Client:                        writer,
			cachedClient:                  nil, // not used in these tests
			scheme:                        scheme,
			objectInformers:               sync.Map{},
			supportedGroupKinds:           sync.Map{},
			additionalSupportedGroupKinds: sync.Map{},
			supportedGvks:                 sync.Map{},
			unexistingGvks:                sync.Map{},
		}
	})

	// Note: tests avoid exercising controller-runtime cache directly to keep them hermetic

	Context("Helpers and utilities", func() {
		It("RemoveDuplicateInformers removes older versions per group-kind", func() {
			var m sync.Map
			var unexisting sync.Map
			gvkOld := makeGVK("apps", "v1alpha1", "Deployment")
			gvkNew := makeGVK("apps", "v2", "Deployment")
			m.Store(gvkOld, &testCachedInformer{synced: true})
			m.Store(gvkNew, &testCachedInformer{synced: true})

			RemoveDuplicateInformers(&m, &unexisting)

			_, okOld := m.Load(gvkOld)
			_, okNew := m.Load(gvkNew)
			Expect(okOld).To(BeFalse())
			Expect(okNew).To(BeTrue())
		})

		It("GetGvksMapPerGroupKind groups by GK from map and sync.Map", func() {
			gvk1 := makeGVK("apps", "v1", "Deployment")
			gvk2 := makeGVK("apps", "v1beta1", "Deployment")
			gvk3 := makeGVK("batch", "v1", "Job")

			// generic map
			m := map[schema.GroupVersionKind]bool{gvk1: true, gvk2: true, gvk3: true}
			grouped := GetGvksMapPerGroupKind(m)
			Expect(grouped).To(HaveLen(2))
			Expect(grouped[gvk1.GroupKind()]).To(ContainElements(gvk1, gvk2))
			Expect(grouped[gvk3.GroupKind()]).To(ContainElements(gvk3))

			// sync.Map version
			var sm sync.Map
			sm.Store(gvk1, true)
			sm.Store(gvk2, true)
			sm.Store(gvk3, true)
			grouped2 := GetGvksMapPerGroupKindFromMap(&sm)
			Expect(grouped2).To(HaveLen(2))
			Expect(grouped2[gvk1.GroupKind()]).To(ContainElements(gvk1, gvk2))
		})

		It("ExternalWorkloadByOriginalWorkloadUIDIndexer returns UID string", func() {
			ew := &runv1alpha1.ExternalWorkload{}
			ew.Spec.OriginalWorkloadUID = types.UID("abc-123")
			res := ExternalWorkloadByOriginalWorkloadUIDIndexer(ew)
			Expect(res).To(Equal([]string{"abc-123"}))
		})

		It("ShouldDropWorkload returns true for configured run.ai kinds", func() {
			dropKinds := []string{"TrainingWorkload", "InteractiveWorkload", "InferenceWorkload", "DistributedInferenceWorkload", "DistributedWorkload"}
			for _, k := range dropKinds {
				obj := makeUnstructured(makeGVK("run.ai", "v1", k))
				Expect(ShouldDropWorkload(obj)).To(BeTrue())
			}
			obj := makeUnstructured(makeGVK("example.com", "v1", "Foo"))
			Expect(ShouldDropWorkload(obj)).To(BeFalse())
		})

		It("AdditionalGroupKind counter increments and decrements correctly", func() {
			gk := schema.GroupKind{Group: "apps", Kind: "Deployment"}
			// starts empty
			_, ok := ewc.additionalSupportedGroupKinds.Load(gk)
			Expect(ok).To(BeFalse())

			ewc.AddAdditionalGroupKind(gk)
			v, ok := ewc.additionalSupportedGroupKinds.Load(gk)
			Expect(ok).To(BeTrue())
			Expect(v.(int)).To(Equal(1))

			ewc.AddAdditionalGroupKind(gk)
			v, _ = ewc.additionalSupportedGroupKinds.Load(gk)
			Expect(v.(int)).To(Equal(2))

			ewc.RemoveAdditionalGroupKind(gk)
			v, _ = ewc.additionalSupportedGroupKinds.Load(gk)
			Expect(v.(int)).To(Equal(1))

			ewc.RemoveAdditionalGroupKind(gk)
			_, ok = ewc.additionalSupportedGroupKinds.Load(gk)
			Expect(ok).To(BeFalse())
		})

		It("GetObjectInformers returns a snapshot of current informers", func() {
			gvk := makeGVK("apps", "v1", "Deployment")
			ewc.objectInformers.Store(gvk, &testCachedInformer{synced: true})
			snap := ewc.GetObjectInformers()
			Expect(snap).To(HaveKey(gvk))
		})
	})
})
