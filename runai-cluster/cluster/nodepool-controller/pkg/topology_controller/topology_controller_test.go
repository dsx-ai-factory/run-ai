package topology_controller

import (
	"context"
	"testing"

	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testKaiTopologyName   = "kai-topo"
	testGroveTopologyName = "grove-topo"
)

func TestKaiTopologyController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KaiTopologyController Suite")
}

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	Expect(kaiv1alpha1.AddToScheme(s)).To(Succeed())
	Expect(grovev1alpha1.AddToScheme(s)).To(Succeed())
	return s
}

func newFakeClient(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}

func newKaiTopology(name string) *kaiv1alpha1.Topology {
	return &kaiv1alpha1.Topology{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newGroveClusterTopologyBinding(name, kaiTopologyName string) *grovev1alpha1.ClusterTopologyBinding {
	ct := &grovev1alpha1.ClusterTopologyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if kaiTopologyName != "" {
		ct.Status.SchedulerTopologyStatuses = []grovev1alpha1.SchedulerTopologyStatus{
			{
				SchedulerTopologyBinding: grovev1alpha1.SchedulerTopologyBinding{
					SchedulerName:     kaiSchedulerName,
					TopologyReference: kaiTopologyName,
				},
			},
		}
	}
	return ct
}

func getKaiTopology(ctx context.Context, c client.Client, name string) *kaiv1alpha1.Topology {
	topology := &kaiv1alpha1.Topology{}
	err := c.Get(ctx, types.NamespacedName{Name: name}, topology)
	Expect(err).ToNot(HaveOccurred())
	return topology
}

var _ = Describe("KaiTopologyController", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(cancel)

		config.SetCurrent(&config.NodePoolControllerConfig{
			GroveTopologyAnnotation:                "kai.scheduler/grove-topology-name",
			GroveTopologyResourceVersionAnnotation: "kai.scheduler/grove-topology-resource-version",
		})
	})

	Context("grove ClusterTopologyBinding created before KAI Topology", func() {
		It("should set annotation when both exist and remove it when grove is deleted", func() {
			groveTopology := newGroveClusterTopologyBinding(testGroveTopologyName, testKaiTopologyName)
			kaiTopology := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, groveTopology, kaiTopology)
			tc := NewKaiTopologyController(fakeClient)

			// Reconcile the KAI topology
			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			// Verify annotations are set
			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).To(HaveKeyWithValue(config.Get().GroveTopologyAnnotation, testGroveTopologyName))
			Expect(updated.Annotations).To(HaveKey(config.Get().GroveTopologyResourceVersionAnnotation))

			// Delete the grove ClusterTopologyBinding
			Expect(fakeClient.Delete(ctx, groveTopology)).To(Succeed())

			// Reconcile again
			_, err = tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			// Verify annotations are removed
			updated = getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyAnnotation))
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyResourceVersionAnnotation))
		})
	})

	Context("KAI Topology created before grove ClusterTopologyBinding", func() {
		It("should not have annotation without grove topology, then set it when grove is created", func() {
			kaiTopology := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, kaiTopology)
			tc := NewKaiTopologyController(fakeClient)

			// Reconcile — no grove topology exists yet
			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			// Verify no annotation
			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyAnnotation))

			// Create grove ClusterTopologyBinding with the mapping annotation
			groveTopology := newGroveClusterTopologyBinding(testGroveTopologyName, testKaiTopologyName)
			Expect(fakeClient.Create(ctx, groveTopology)).To(Succeed())

			// Reconcile again — now the grove topology exists
			_, err = tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			// Verify annotations are now set
			updated = getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).To(HaveKeyWithValue(config.Get().GroveTopologyAnnotation, testGroveTopologyName))
			Expect(updated.Annotations).To(HaveKey(config.Get().GroveTopologyResourceVersionAnnotation))
		})
	})

	Context("KAI Topology not found", func() {
		It("should not error when reconciling a missing KAI Topology", func() {
			scheme := newScheme()
			fakeClient := newFakeClient(scheme)
			tc := NewKaiTopologyController(fakeClient)

			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("grove ClusterTopologyBinding has no kai-scheduler status entry", func() {
		It("should not set annotations on the KAI Topology", func() {
			groveTopology := &grovev1alpha1.ClusterTopologyBinding{
				ObjectMeta: metav1.ObjectMeta{Name: testGroveTopologyName},
				Status: grovev1alpha1.ClusterTopologyBindingStatus{
					SchedulerTopologyStatuses: []grovev1alpha1.SchedulerTopologyStatus{
						{
							SchedulerTopologyBinding: grovev1alpha1.SchedulerTopologyBinding{
								SchedulerName:     "other-scheduler",
								TopologyReference: testKaiTopologyName,
							},
						},
					},
				},
			}
			kaiTopology := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, groveTopology, kaiTopology)
			tc := NewKaiTopologyController(fakeClient)

			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyAnnotation))
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyResourceVersionAnnotation))
		})
	})

	Context("grove ClusterTopologyBinding references a different KAI Topology", func() {
		It("should not annotate an unrelated KAI Topology", func() {
			groveTopology := newGroveClusterTopologyBinding(testGroveTopologyName, "other-kai-topo")
			unrelatedKai := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, groveTopology, unrelatedKai)
			tc := NewKaiTopologyController(fakeClient)

			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyAnnotation))
			Expect(updated.Annotations).ToNot(HaveKey(config.Get().GroveTopologyResourceVersionAnnotation))
		})
	})

	Context("multiple grove ClusterTopologyBindings are present", func() {
		It("should annotate the KAI Topology with the matching grove name", func() {
			matching := newGroveClusterTopologyBinding("grove-match", testKaiTopologyName)
			other := newGroveClusterTopologyBinding("grove-other", "other-kai-topo")
			kaiTopology := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, matching, other, kaiTopology)
			tc := NewKaiTopologyController(fakeClient)

			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations).To(HaveKeyWithValue(config.Get().GroveTopologyAnnotation, "grove-match"))
		})
	})

	Context("grove ClusterTopologyBinding resourceVersion changes", func() {
		It("should be idempotent on no-op reconciles and update RV when grove changes", func() {
			groveTopology := newGroveClusterTopologyBinding(testGroveTopologyName, testKaiTopologyName)
			kaiTopology := newKaiTopology(testKaiTopologyName)

			scheme := newScheme()
			fakeClient := newFakeClient(scheme, groveTopology, kaiTopology)
			tc := NewKaiTopologyController(fakeClient)

			// First reconcile: annotations are set
			_, err := tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated := getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			firstGroveRV := updated.Annotations[config.Get().GroveTopologyResourceVersionAnnotation]
			Expect(firstGroveRV).ToNot(BeEmpty())
			kaiRVAfterFirst := updated.ResourceVersion

			// Second reconcile with nothing changed: KAI topology should not be re-patched
			_, err = tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated = getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.ResourceVersion).To(Equal(kaiRVAfterFirst))

			// Bump the grove resource version
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: testGroveTopologyName}, groveTopology)).To(Succeed())
			if groveTopology.Labels == nil {
				groveTopology.Labels = map[string]string{}
			}
			groveTopology.Labels["touch"] = "1"
			Expect(fakeClient.Update(ctx, groveTopology)).To(Succeed())

			// Third reconcile: annotation RV is updated
			_, err = tc.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
			})
			Expect(err).ToNot(HaveOccurred())

			updated = getKaiTopology(ctx, fakeClient, testKaiTopologyName)
			Expect(updated.Annotations[config.Get().GroveTopologyResourceVersionAnnotation]).ToNot(Equal(firstGroveRV))
		})
	})
})

var _ = Describe("groveToKaiTopologyName", func() {
	It("returns the kai-scheduler entry's TopologyReference", func() {
		ct := newGroveClusterTopologyBinding(testGroveTopologyName, testKaiTopologyName)
		Expect(groveToKaiTopologyName(ct)).To(Equal(testKaiTopologyName))
	})

	It("returns empty when there is no kai-scheduler entry", func() {
		ct := &grovev1alpha1.ClusterTopologyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: testGroveTopologyName},
			Status: grovev1alpha1.ClusterTopologyBindingStatus{
				SchedulerTopologyStatuses: []grovev1alpha1.SchedulerTopologyStatus{
					{
						SchedulerTopologyBinding: grovev1alpha1.SchedulerTopologyBinding{
							SchedulerName:     "other-scheduler",
							TopologyReference: "other-topo",
						},
					},
				},
			},
		}
		Expect(groveToKaiTopologyName(ct)).To(BeEmpty())
	})

	It("returns empty when Status is empty", func() {
		ct := &grovev1alpha1.ClusterTopologyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: testGroveTopologyName},
		}
		Expect(groveToKaiTopologyName(ct)).To(BeEmpty())
	})

	It("returns empty for a non-ClusterTopologyBinding object", func() {
		Expect(groveToKaiTopologyName(&kaiv1alpha1.Topology{})).To(BeEmpty())
	})
})

var _ = Describe("mapGroveToKaiReconcileRequest", func() {
	var tc *KaiTopologyController

	BeforeEach(func() {
		tc = NewKaiTopologyController(newFakeClient(newScheme()))
	})

	It("returns a reconcile request for the mapped KAI Topology", func() {
		ct := newGroveClusterTopologyBinding(testGroveTopologyName, testKaiTopologyName)
		reqs := tc.mapGroveToKaiReconcileRequest(context.Background(), ct)
		Expect(reqs).To(ConsistOf(ctrl.Request{
			NamespacedName: types.NamespacedName{Name: testKaiTopologyName},
		}))
	})

	It("returns nil when the grove topology has no kai-scheduler entry", func() {
		ct := &grovev1alpha1.ClusterTopologyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: testGroveTopologyName},
		}
		Expect(tc.mapGroveToKaiReconcileRequest(context.Background(), ct)).To(BeNil())
	})
})
