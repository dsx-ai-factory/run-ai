package kartacontroller

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/kartacontroller/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test/fakecachedclient"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKartaController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KartaController Suite")
}

var _ = Describe("KartaController (Karta-centric)", Ordered, func() {
	var (
		scheme             *runtime.Scheme
		k8s                client.WithWatch
		rc                 *KartaController
		ctx                context.Context
		stopper            chan struct{}
		kartaControlParams KartaControlParams
		fakeCachedClient   *fakecachedclient.FakeCachedClient

		exitCalled bool
		oldExit    func(int) // stub for exit function to avoid exiting tests
	)

	BeforeAll(func() {
		scheme = buildTestScheme()
		ctx = context.Background()
		stopper = make(chan struct{}, 1)
		DeferCleanup(func() {
			close(stopper)
		})

		fakeCachedClient = fakecachedclient.NewFakeCachedClient(k8s, stopper, true)

		// stub exitFn to avoid exiting tests and capture calls
		oldExit = exitFn
		exitFn = func(code int) { exitCalled = true }
	})

	BeforeEach(func() {
		k8s = fake.NewClientBuilder().WithScheme(scheme).
			WithStatusSubresource(&kartav1alpha1.Karta{}).Build()
		kartaControlParams = KartaControlParams{
			RBACUpdaterEnabled:       false,
			RegisteredAdditionalGVKs: []schema.GroupVersionKind{},
			CachedAdditionalGvksApi:  fakeCachedClient,
		}
		rc = NewKartaController(k8s, &kartaControlParams)

		exitCalled = false
	})

	AfterAll(func() {
		exitFn = oldExit
	})

	Context("Reconcile with Karta object", func() {
		It("error when RootComponent.Kind is nil", func() {
			karta := buildKarta("karta-no-gvk", nil)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))
			// CRDExists / KartaValidated are owned by the OSS controller and not set by EWI.
			_, hasCRD := tryFindCond(karta.Status.Conditions, string(runv1alpha1.CRDExists))
			_, hasValidated := tryFindCond(karta.Status.Conditions, string(runv1alpha1.KartaValidated))
			Expect(hasCRD).To(BeFalse())
			Expect(hasValidated).To(BeFalse())
			Expect(exitCalled).To(BeFalse())
		})

		It("sets RBACReady=true, SyncEnabled=true when kind is already supported; RunaiReady stays false without the OSS Ready", func() {
			dGVK := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			karta := buildKarta("karta-supported", &dGVK)
			expectCreate(ctx, k8s, karta)
			// OSS controller reports the CRD exists.
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(exitCalled).To(BeFalse())
		})

		It("skips RBAC and does not restart when the operator has not reported the CRD as existing", func() {
			dGVK := schema.GroupVersionKind{Group: "bla", Version: "v1", Kind: "Lies"}
			karta := buildKarta("karta-missing-crd", &dGVK)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(exitCalled).To(BeFalse())
		})

		It("sets RunaiReady=true when the OSS Ready condition is True and RBAC is ready", func() {
			dGVK := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			karta := buildKarta("karta-valid", &dGVK)
			expectCreate(ctx, k8s, karta)
			// Simulate the OSS Karta controller having validated + marked the Karta Ready.
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.Ready, metav1.ConditionTrue)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.Ready).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RunaiReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(exitCalled).To(BeFalse())
		})

		It("restarts when the OSS controller reports the CRD exists but the GVK is not yet registered", func() {
			barGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
			karta := buildKarta("karta-bar-unregistered", &barGVK)
			expectCreate(ctx, k8s, karta)
			// OSS controller has confirmed the CRD exists.
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(exitCalled).To(BeTrue())
		})

		It("does not restart when the OSS controller has not reported the CRD as existing", func() {
			barGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
			karta := buildKarta("karta-bar-no-crd", &barGVK)
			expectCreate(ctx, k8s, karta)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(exitCalled).To(BeFalse())
		})

		It("error nil when karta doesn't exist", func() {
			_, err := rc.Reconcile(ctx, getRequestFromName("karta-missing"))
			Expect(err).To(BeNil())
			Expect(exitCalled).To(BeFalse())
		})
	})

	Context("Reconcile RBAC integration", func() {
		It("sets RBACReady=false and returns error when RBAC update fails", func() {
			// CRD exists for unsupported kind
			createCRD(ctx, k8s, "foos.test.run.ai", "test.run.ai", "Foo", []string{"v1"})
			fooGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Foo"}
			karta := buildKarta("karta-foo", &fooGVK)
			expectCreate(ctx, k8s, karta)
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)
			// updater disabled; RBAC missing -> error
			kartaControlParams.ClusterRoleReadName = "role-disabled"
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(HaveOccurred())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))

			Expect(exitCalled).To(BeFalse())
		})

		It("sets RBACReady=true when RBAC update succeeds and writes rule to ClusterRole", func() {
			// CRD exists for unsupported kind
			createCRD(ctx, k8s, "bars.test.run.ai", "test.run.ai", "Bar", []string{"v1"})
			barGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
			karta := buildKarta("karta-bar", &barGVK)
			karta.Spec.StructureDefinition.AdditionalChildKinds = []kartav1alpha1.GroupVersionKind{
				{Group: "test.run.ai", Version: "v1", Kind: "Baz"},
				{Group: "other.run.ai", Version: "v1", Kind: "Qux"},
			}
			expectCreate(ctx, k8s, karta)
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)

			// prepare empty ClusterRole and enable updater
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-enabled"}}
			expectCreate(ctx, k8s, cr)
			kartaControlParams.RBACUpdaterEnabled = true
			kartaControlParams.ClusterRoleReadName = cr.Name
			kartaControlParams.ClusterRoleCrudName = "role-enabled-crud"
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))

			// verify rule presence on clusterrole
			updatedCR := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: cr.Name}, updatedCR)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCR, "test.run.ai", "bars", false)).To(BeTrue())
			updatedCRCrud := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: "role-enabled-crud"}, updatedCRCrud)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, "test.run.ai", "bars", true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, "test.run.ai", "bazs", true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, "other.run.ai", "quxs", true)).To(BeTrue())
			Expect(exitCalled).To(BeFalse())

			additionalGvks := fakeCachedClient.GetAdditionalGroupKinds()
			Expect(additionalGvks[schema.GroupKind{Group: "test.run.ai", Kind: "Baz"}]).To(Equal(1))
			Expect(additionalGvks[schema.GroupKind{Group: "other.run.ai", Kind: "Qux"}]).To(Equal(1))
		})

		It("sets SyncEnabled=false and restarts when gvk not registered and the CRD exists", func() {
			createCRD(ctx, k8s, "bars.test.run.ai", "test.run.ai", "Bar", []string{"v1"})
			barGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
			karta := buildKarta("karta-bar", &barGVK)
			expectCreate(ctx, k8s, karta)
			// OSS controller has confirmed the CRD exists, so a restart can register the GVK.
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-enabled"}}
			expectCreate(ctx, k8s, cr)

			kartaControlParams.RBACUpdaterEnabled = true
			kartaControlParams.ClusterRoleReadName = "role-enabled"
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{}
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(exitCalled).To(BeTrue())
		})

		It("sets SyncEnabled=true when all succeeds and gvk configured", func() {
			// CRD exists for unsupported kind
			createCRD(ctx, k8s, "bars.test.run.ai", "test.run.ai", "Bar", []string{"v1"})
			barGVK := schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
			karta := buildKarta("karta-bar", &barGVK)
			expectCreate(ctx, k8s, karta)
			seedKartaCondition(ctx, k8s, karta, runv1alpha1.CRDExists, metav1.ConditionTrue)
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-enabled"}}
			expectCreate(ctx, k8s, cr)

			kartaControlParams.RBACUpdaterEnabled = true
			kartaControlParams.ClusterRoleReadName = "role-enabled"
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACReady).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncEnabled).Status).To(Equal(metav1.ConditionTrue))
			Expect(exitCalled).To(BeFalse())
		})
	})

	Context("Deletion flow", func() {
		var (
			barGVK = schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Bar"}
		)

		BeforeEach(func() {
			createCRD(ctx, k8s, barGVK.Kind, barGVK.Group, barGVK.Kind, []string{"v1"})

			kartaControlParams.InformersApi = fakeCachedClient
			kartaControlParams.ClusterRoleReadName = ""
			kartaControlParams.ClusterRoleCrudName = ""
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			fakeCachedClient.SetInformers(ctx, []schema.GroupVersionKind{}, []schema.GroupVersionKind{barGVK})
			informers := fakeCachedClient.GetObjectInformers()
			_, informerFound := informers[barGVK]
			Expect(informerFound).To(BeTrue())
		})

		It("skips finalization when controller is not in finalizers", func() {
			karta := buildKarta("karta-del-skip", &barGVK)
			karta.Finalizers = []string{"some.other/finalizer"}
			expectCreate(ctx, k8s, karta)
			Expect(k8s.Delete(ctx, karta)).To(Succeed())

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(strSliceContains(karta.Finalizers, FinalizerName)).To(BeFalse())
			Expect(exitCalled).To(BeFalse())
			expectHaveInformer(fakeCachedClient, barGVK, true)
			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeTrue())
		})

		It("removes finalizer, unregisters GVK and removes informer on successful deletion", func() {
			karta := buildKarta("karta-del-success", &barGVK)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())
			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(strSliceContains(karta.Finalizers, FinalizerName)).To(BeTrue())

			Expect(k8s.Delete(ctx, karta)).To(Succeed())
			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)

			_, err = rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			Expect(k8s.Get(ctx, getObjectKeyForObject(karta), karta)).To(HaveOccurred())
			expectHaveInformer(fakeCachedClient, barGVK, false)
			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeFalse())
		})

		It("updates status and returns error when ExternalWorkloads still uses the GVK", func() {
			karta := buildKarta("karta-del-ew-exist", &barGVK)
			karta.Finalizers = []string{FinalizerName}
			expectCreate(ctx, k8s, karta)

			ew := &runv1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{Name: "ew-using-baz"},
				Spec:       runv1alpha1.ExternalWorkloadSpec{Group: barGVK.Group, Type: barGVK.Kind},
			}
			expectCreate(ctx, k8s, ew)

			Expect(k8s.Delete(ctx, karta)).To(Succeed())

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(HaveOccurred())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncDisabled).Status).To(Equal(metav1.ConditionTrue))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACCleanupSuccess).Status).To(Equal(metav1.ConditionFalse))
			Expect(strSliceContains(karta.Finalizers, FinalizerName)).To(BeTrue())
			expectHaveInformer(fakeCachedClient, barGVK, false)
			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeTrue())
		})

		It("updates status and returns success when ClusterRole doesn't exist", func() {
			karta := buildKarta("karta-del-rbac-fail", &barGVK)
			karta.Finalizers = []string{FinalizerName}
			expectCreate(ctx, k8s, karta)
			Expect(k8s.Delete(ctx, karta)).To(Succeed())

			kartaControlParams.ClusterRoleReadName = "non-existent-role"
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			Expect(k8s.Get(ctx, getObjectKeyForObject(karta), karta)).To(HaveOccurred())
			expectHaveInformer(fakeCachedClient, barGVK, false)
			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeFalse())
		})

		It("updates status and returns error when informer removal fails", func() {
			karta := buildKarta("karta-del-informer-fail", &barGVK)
			karta.Finalizers = []string{FinalizerName}
			expectCreate(ctx, k8s, karta)
			Expect(k8s.Delete(ctx, karta)).To(Succeed())

			fakeIR := &fakeInformerRemover{err: fmt.Errorf("informer removal simulated error")}
			kartaControlParams.InformersApi = fakeIR
			rc = NewKartaController(k8s, &kartaControlParams)

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(HaveOccurred())

			expectGet(ctx, k8s, getObjectKeyForObject(karta), karta)
			Expect(findCond(karta.Status.Conditions, runv1alpha1.SyncDisabled).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionFalse))
			Expect(findCond(karta.Status.Conditions, runv1alpha1.RBACCleanupSuccess).Status).To(Equal(metav1.ConditionFalse))
			Expect(strSliceContains(karta.Finalizers, FinalizerName)).To(BeTrue())

			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeTrue())
			expectHaveInformer(fakeCachedClient, barGVK, true)
			Expect(len(fakeIR.calls)).To(Equal(1))
			Expect(fakeIR.calls[0]).To(Equal(barGVK))
		})

		It("removes finalizer when GVK is not defined (nil)", func() {
			karta := buildKarta("karta-del-no-gvk", nil)
			karta.Finalizers = []string{FinalizerName}
			expectCreate(ctx, k8s, karta)
			Expect(k8s.Delete(ctx, karta)).To(Succeed())

			_, err := rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			Expect(k8s.Get(ctx, getObjectKeyForObject(karta), karta)).To(HaveOccurred())
		})

		It("validate RBAC removed after deletion when it was the last Karta for the GK", func() {
			// prepare ClusterRoles with read/crud rule
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-enabled"}}
			rule := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{"get", "list", "watch"}}
			cr.Rules = append(cr.Rules, rule)
			expectCreate(ctx, k8s, cr)
			crCrud := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-enabled-crud"}}
			ruleCrud := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{"get", "list", "watch", "delete", "update", "patch", "create"}}
			crCrud.Rules = append(cr.Rules, ruleCrud)
			expectCreate(ctx, k8s, crCrud)

			// enable RBAC removal and keep only this GVK registered
			kartaControlParams.RBACUpdaterEnabled = true
			kartaControlParams.ClusterRoleReadName = cr.Name
			kartaControlParams.ClusterRoleCrudName = crCrud.Name
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			// Karta lifecycle: create, add finalizer, delete, finalize
			karta := buildKarta("karta-last-gk", &barGVK)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta)) // add finalizer
			Expect(err).To(BeNil())

			Expect(k8s.Delete(ctx, karta)).To(Succeed())
			_, err = rc.Reconcile(ctx, getRequestForObj(karta)) // finalize
			Expect(err).To(BeNil())

			// Karta finalized and informer removed
			Expect(k8s.Get(ctx, getObjectKeyForObject(karta), karta)).To(HaveOccurred())
			expectHaveInformer(fakeCachedClient, barGVK, false)
			Expect(rc.isGvkRegisteredAsAdditional(barGVK)).To(BeFalse())

			// RBAC rule cleaned up (last Karta for this GK)
			updatedCR := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: cr.Name}, updatedCR)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCR, barGVK.Group, "bars", false)).To(BeFalse())
			updatedCRCrud := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: crCrud.Name}, updatedCRCrud)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, barGVK.Group, "bars", true)).To(BeFalse())
		})

		It("updates status and returns success when RBACUpdaterEnabled is false", func() {
			// prepare ClusterRoles with read/crud rule that needs removal
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-disabled"}}
			rule := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{"get", "list", "watch"}}
			cr.Rules = append(cr.Rules, rule)
			expectCreate(ctx, k8s, cr)
			crCrud := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-disabled-crud"}}
			ruleCrud := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{"get", "list", "watch", "delete", "update", "patch", "create"}}
			crCrud.Rules = append(cr.Rules, ruleCrud)
			expectCreate(ctx, k8s, crCrud)

			// disable RBAC removal but set cluster role name so removal is required
			kartaControlParams.RBACUpdaterEnabled = false
			kartaControlParams.ClusterRoleReadName = cr.Name
			kartaControlParams.ClusterRoleCrudName = crCrud.Name
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			// Karta lifecycle to trigger finalize
			karta := buildKarta("karta-last-gk-rbac-disabled", &barGVK)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta)) // add finalizer
			Expect(err).To(BeNil())

			Expect(k8s.Delete(ctx, karta)).To(Succeed())
			_, err = rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())

			// RBAC rule still exists - the client user will have to delete if they want
			updatedCR := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: cr.Name}, updatedCR)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCR, barGVK.Group, "bars", false)).To(BeTrue())
			updatedCRCrud := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: crCrud.Name}, updatedCRCrud)).To(Succeed())
			Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, barGVK.Group, "bars", true)).To(BeTrue())
		})

		It("updates status and returns success when cluster role already doesn't have permissions", func() {
			// prepare ClusterRoles with partial read/crud rules that don't need removal
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-disabled"}}
			rule := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{}}
			cr.Rules = append(cr.Rules, rule)
			expectCreate(ctx, k8s, cr)
			crCrud := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "role-disabled-crud"}}
			ruleCrud := rbacv1.PolicyRule{APIGroups: []string{barGVK.Group}, Resources: []string{"bars"}, Verbs: []string{"get", "list", "create"}}
			crCrud.Rules = append(cr.Rules, ruleCrud)
			expectCreate(ctx, k8s, crCrud)

			kartaControlParams.RBACUpdaterEnabled = true
			kartaControlParams.ClusterRoleReadName = cr.Name
			kartaControlParams.ClusterRoleCrudName = crCrud.Name
			kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{barGVK}
			rc = NewKartaController(k8s, &kartaControlParams)

			// Karta lifecycle to trigger finalize
			karta := buildKarta("karta-last-gk-rbac-removed", &barGVK)
			expectCreate(ctx, k8s, karta)
			_, err := rc.Reconcile(ctx, getRequestForObj(karta)) // add finalizer
			Expect(err).To(BeNil())

			Expect(k8s.Delete(ctx, karta)).To(Succeed())
			_, err = rc.Reconcile(ctx, getRequestForObj(karta))
			Expect(err).To(BeNil())
		})

		Context("several versions for same GV", func() {
			var (
				fooGVK = schema.GroupVersionKind{Group: "test.run.ai", Version: "v1", Kind: "Foo"}
				gvkV1  = schema.GroupVersionKind{Group: fooGVK.Group, Version: "v1", Kind: fooGVK.Kind}
				gvkV2  = schema.GroupVersionKind{Group: fooGVK.Group, Version: "v2", Kind: fooGVK.Kind}

				clusterRoleName     = "role-enabled"
				clusterRoleCrudName = "role-enabled-crud"

				kartaV1, kartaV2 *kartav1alpha1.Karta
			)

			BeforeEach(func() {
				createCRD(ctx, k8s, fooGVK.Kind, fooGVK.Group, fooGVK.Kind, []string{"v1", "v2"})

				// Prepare RBAC with an initial read/crud rule present for this GK
				cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleName}}
				rule := rbacv1.PolicyRule{APIGroups: []string{fooGVK.Group}, Resources: []string{"foos"}, Verbs: []string{"get", "list", "watch"}}
				cr.Rules = append(cr.Rules, rule)
				expectCreate(ctx, k8s, cr)
				crCrud := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: clusterRoleCrudName}}
				ruleCrud := rbacv1.PolicyRule{APIGroups: []string{fooGVK.Group}, Resources: []string{"foos"}, Verbs: []string{"get", "list", "watch", "delete", "update", "patch", "create"}}
				crCrud.Rules = append(cr.Rules, ruleCrud)
				expectCreate(ctx, k8s, crCrud)

				// Configure controller with both versions registered and RBAC enabled for deletion
				kartaControlParams.RBACUpdaterEnabled = true
				kartaControlParams.ClusterRoleReadName = cr.Name
				kartaControlParams.ClusterRoleCrudName = crCrud.Name
				kartaControlParams.RegisteredAdditionalGVKs = []schema.GroupVersionKind{gvkV1, gvkV2}
				rc = NewKartaController(k8s, &kartaControlParams)
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeTrue())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeTrue())

				// Set informers for both versions of the same GK
				fakeCachedClient.SetInformers(ctx, []schema.GroupVersionKind{}, []schema.GroupVersionKind{gvkV1, gvkV2})
				// should only have the latest version informer
				expectHaveInformer(fakeCachedClient, gvkV1, false)
				expectHaveInformer(fakeCachedClient, gvkV2, true)

				// create 2 Kartas, one for each version
				kartaV1 = buildKarta("karta-del-gvk-v1", &gvkV1)
				kartaV2 = buildKarta("karta-del-gvk-v2", &gvkV2)
				expectCreate(ctx, k8s, kartaV1)
				_, err := rc.Reconcile(ctx, getRequestForObj(kartaV1)) // add finalizer
				Expect(err).To(BeNil())
				expectCreate(ctx, k8s, kartaV2)
				_, err = rc.Reconcile(ctx, getRequestForObj(kartaV2)) // add finalizer
				Expect(err).To(BeNil())
			})

			It("workloads exist for GK, 1st version deletion should succeed, 2nf version should fail and keep finalizer", func() {
				// Create EW for the GK
				ew := &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Name: "ew-using-gk"}, Spec: runv1alpha1.ExternalWorkloadSpec{Group: fooGVK.Group, Type: fooGVK.Kind}}
				expectCreate(ctx, k8s, ew)

				// should succeed to delete kartaV2 - although it is the one with main informer, will move the informer to v1
				Expect(k8s.Delete(ctx, kartaV2)).To(Succeed())
				_, err := rc.Reconcile(ctx, getRequestForObj(kartaV2))
				Expect(err).ToNot(HaveOccurred())
				Expect(k8s.Get(ctx, getObjectKeyForObject(kartaV2), kartaV2)).To(HaveOccurred())
				// Finalizer still present, informer for v2 removed, informer for v1 added
				expectHaveInformer(fakeCachedClient, gvkV2, false)
				expectHaveInformer(fakeCachedClient, gvkV1, true)
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeFalse())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeTrue())

				// RBAC rule should still exist because there is another version with informer for the same GK
				updatedCR := &rbacv1.ClusterRole{}
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleName}, updatedCR)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCR, fooGVK.Group, "foos", false)).To(BeTrue())
				updatedCRCrud := &rbacv1.ClusterRole{}
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleCrudName}, updatedCRCrud)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, fooGVK.Group, "foos", true)).To(BeTrue())

				// try to delete kartaV1 - should fail because EW exists
				Expect(k8s.Delete(ctx, kartaV1)).To(Succeed())
				_, err = rc.Reconcile(ctx, getRequestForObj(kartaV1))
				Expect(err).To(HaveOccurred())
				expectGet(ctx, k8s, getObjectKeyForObject(kartaV1), kartaV1)
				Expect(findCond(kartaV1.Status.Conditions, runv1alpha1.SyncDisabled).Status).To(Equal(metav1.ConditionTrue))
				Expect(findCond(kartaV1.Status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionFalse))
				Expect(findCond(kartaV1.Status.Conditions, runv1alpha1.RBACCleanupSuccess).Status).To(Equal(metav1.ConditionFalse))
				Expect(strSliceContains(kartaV1.Finalizers, FinalizerName)).To(BeTrue())
				// Informer for v1 should not be present
				expectHaveInformer(fakeCachedClient, gvkV2, false)
				expectHaveInformer(fakeCachedClient, gvkV1, false)
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeFalse())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeTrue())

				// Cleanup EW and retry deletion – this time it should proceed
				Expect(k8s.Delete(ctx, ew)).To(Succeed())
				_, err = rc.Reconcile(ctx, getRequestForObj(kartaV1)) // finalize
				Expect(err).To(BeNil())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeFalse())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeFalse())

				// After finalize, v1+v2 informers removed; RBAC should be removed
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleName}, updatedCR)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCR, fooGVK.Group, "foos", false)).To(BeFalse())
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleCrudName}, updatedCRCrud)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, fooGVK.Group, "foos", true)).To(BeFalse())
			})

			It("workloads exist for GK different version, deletion should succeed", func() {
				// Create EW for the GK
				ew := &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Name: "ew-using-gk"}, Spec: runv1alpha1.ExternalWorkloadSpec{Group: fooGVK.Group, Type: fooGVK.Kind}}
				expectCreate(ctx, k8s, ew)

				// should succeed to delete kartaV1 - it doesn't have the informer anyway
				Expect(k8s.Delete(ctx, kartaV1)).To(Succeed())
				_, err := rc.Reconcile(ctx, getRequestForObj(kartaV1))
				Expect(err).ToNot(HaveOccurred())
				Expect(k8s.Get(ctx, getObjectKeyForObject(kartaV1), kartaV1)).To(HaveOccurred())
				expectHaveInformer(fakeCachedClient, gvkV2, true)
				expectHaveInformer(fakeCachedClient, gvkV1, false)
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeTrue())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeFalse())

				// validate ClusterRoles still have permissions because v2 informer still exists
				updatedCR := &rbacv1.ClusterRole{}
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleName}, updatedCR)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCR, fooGVK.Group, "foos", false)).To(BeTrue())
				updatedCRCrud := &rbacv1.ClusterRole{}
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleCrudName}, updatedCRCrud)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, fooGVK.Group, "foos", true)).To(BeTrue())

				// try to delete v2 - should fail because EW exists
				Expect(k8s.Delete(ctx, kartaV2)).To(Succeed())
				_, err = rc.Reconcile(ctx, getRequestForObj(kartaV2))
				Expect(err).To(HaveOccurred())
				expectGet(ctx, k8s, getObjectKeyForObject(kartaV2), kartaV2)
				Expect(findCond(kartaV2.Status.Conditions, runv1alpha1.SyncDisabled).Status).To(Equal(metav1.ConditionTrue))
				Expect(findCond(kartaV2.Status.Conditions, runv1alpha1.NoWorkloadsUseKarta).Status).To(Equal(metav1.ConditionFalse))
				Expect(findCond(kartaV2.Status.Conditions, runv1alpha1.RBACCleanupSuccess).Status).To(Equal(metav1.ConditionFalse))
				Expect(strSliceContains(kartaV2.Finalizers, FinalizerName)).To(BeTrue())
				// informer for v2 should have been removed
				expectHaveInformer(fakeCachedClient, gvkV2, false)
				expectHaveInformer(fakeCachedClient, gvkV1, false)
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeTrue())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeFalse())

				// delete the EW and delete kartaV2 – this time it should proceed
				Expect(k8s.Delete(ctx, ew)).To(Succeed())
				_, err = rc.Reconcile(ctx, getRequestForObj(kartaV2)) // finalize
				Expect(err).To(BeNil())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV2)).To(BeFalse())
				Expect(rc.isGvkRegisteredAsAdditional(gvkV1)).To(BeFalse())
				expectHaveInformer(fakeCachedClient, gvkV2, false)
				expectHaveInformer(fakeCachedClient, gvkV1, false)

				// RBAC rule should be cleaned up
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleName}, updatedCR)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCR, fooGVK.Group, "foos", false)).To(BeFalse())
				Expect(k8s.Get(ctx, client.ObjectKey{Name: clusterRoleCrudName}, updatedCRCrud)).To(Succeed())
				Expect(hasAnyRuleWithAllVerbs(updatedCRCrud, fooGVK.Group, "foos", true)).To(BeFalse())
			})
		})
	})

})

type fakeInformerRemover struct {
	calls []schema.GroupVersionKind
	err   error
}

func (f *fakeInformerRemover) RemoveInformer(_ context.Context, gvk schema.GroupVersionKind) error {
	f.calls = append(f.calls, gvk)
	return f.err
}

func (f *fakeInformerRemover) GetObjectInformers() map[schema.GroupVersionKind]cachedclient.CachedInformer {
	result := make(map[schema.GroupVersionKind]cachedclient.CachedInformer)
	for _, gvk := range f.calls {
		result[gvk] = nil
	}
	return result
}

func buildTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	Expect(corev1.AddToScheme(s)).To(Succeed())
	Expect(rbacv1.AddToScheme(s)).To(Succeed())
	Expect(apiextensionsv1.AddToScheme(s)).To(Succeed())
	Expect(kartav1alpha1.AddToScheme(s)).To(Succeed())
	Expect(runv1alpha1.AddToScheme(s)).To(Succeed())
	return s
}

func buildKarta(name string, kind *schema.GroupVersionKind) *kartav1alpha1.Karta {
	karta := &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	if kind != nil {
		karta.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
			Group:   kind.Group,
			Version: kind.Version,
			Kind:    kind.Kind,
		}
	}
	return karta
}

// seedKartaCondition simulates the OSS Karta controller writing a condition to the
// Karta status subresource.
func seedKartaCondition(ctx context.Context, c client.Client, karta *kartav1alpha1.Karta, condType runv1alpha1.KartaConditionType, status metav1.ConditionStatus) {
	karta.Status.Conditions = append(karta.Status.Conditions, metav1.Condition{
		Type:   string(condType),
		Status: status,
		Reason: string(condType),
	})
	Expect(c.Status().Update(ctx, karta)).To(Succeed())
}

func expectCreate(ctx context.Context, c client.Client, obj client.Object) {
	Expect(c.Create(ctx, obj)).To(Succeed())
}

func expectGet(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	ExpectWithOffset(1, c.Get(ctx, key, obj)).To(Succeed())
}

func createCRD(ctx context.Context, c client.Client, name, group, kind string, versions []string) {
	v := make([]apiextensionsv1.CustomResourceDefinitionVersion, 0, len(versions))
	for _, ver := range versions {
		v = append(v, apiextensionsv1.CustomResourceDefinitionVersion{Name: ver, Served: true, Storage: true})
	}
	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:    group,
			Names:    apiextensionsv1.CustomResourceDefinitionNames{Kind: kind, Plural: toPlural(kind)},
			Scope:    apiextensionsv1.NamespaceScoped,
			Versions: v,
		},
	}
	expectCreate(ctx, c, crd)
}

func toPlural(kind string) string {
	if len(kind) == 0 {
		return ""
	}
	return strings.ToLower(kind) + "s"
}

func tryFindCond(conds []metav1.Condition, t string) (metav1.Condition, bool) {
	for _, c := range conds {
		if c.Type == t {
			return c, true
		}
	}
	return metav1.Condition{}, false
}

func findCond(conds []metav1.Condition, t runv1alpha1.KartaConditionType) metav1.Condition {
	c, ok := tryFindCond(conds, string(t))
	if !ok {
		Fail("condition not found: " + string(t))
	}
	return c
}

func hasAnyRuleWithAllVerbs(cr *rbacv1.ClusterRole, group, resource string, hasCrud bool) bool {
	for _, r := range cr.Rules {
		if strSliceContains(r.APIGroups, group) && strSliceContains(r.Resources, resource) && handlers.HasAllVerbs(r, hasCrud) {
			return true
		}
	}
	return false
}

func strSliceContains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

func getRequestFromName(name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Name: name}}
}

func getRequestForObj(obj client.Object) ctrl.Request {
	return getRequestFromName(obj.GetName())
}

func getObjectKeyForObject(obj client.Object) client.ObjectKey {
	return client.ObjectKey{Name: obj.GetName()}
}

func expectHaveInformer(informerApi handlers.InformersApi, gvk schema.GroupVersionKind, expectFound bool) {
	informers := informerApi.GetObjectInformers()
	_, found := informers[gvk]
	ExpectWithOffset(1, found).To(Equal(expectFound))
}
