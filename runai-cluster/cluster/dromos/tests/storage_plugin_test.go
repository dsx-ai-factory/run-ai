package controller

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/storage"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ = Describe("Storage Plugin Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager            ctrl.Manager
		k8sClient             client.Client
		storageController     *controller.WorkloadRunnerController
		managerCtx            context.Context
		managerCancel         context.CancelFunc
		storageNamespace      = "runai-storage-test"
		storageControllerName = "storage-plugin-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		storagePlugin := storage.NewPlugin(false)
		Expect(storagePlugin).NotTo(BeNil())

		registry := framework.NewRegistry(storagePlugin)
		fw := framework.NewFramework(registry, nil)
		storageController = controller.NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(storageControllerName)).WithName(storageControllerName)
		err = storageController.SetupWithManager(managerCtx, k8sManager)
		Expect(err).NotTo(HaveOccurred())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		startManager(k8sManager, managerCtx, managerCancel)

		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: storageNamespace,
			},
		}
		_ = k8sClient.Create(managerCtx, ns)

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	Context("WorkloadRunner without storage instructions", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		DescribeTable("should set plugin phase to Disabled and create no PVCs",
			func(instructions *dromosv1alpha1.Instructions) {
				suffix := rand.String(4)
				jobset := test.NewJobSetBuilder("test-workload-"+suffix, storageNamespace).WithDefault().Build()

				_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner := createWorkloadRunner(managerCtx, k8sClient, "test-runner-"+suffix, storageNamespace, jobset, instructions)

				updated := &dromosv1alpha1.WorkloadRunner{}
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{
						Name:      runner.Name,
						Namespace: runner.Namespace,
					}, updated)).To(Succeed())
					g.Expect(updated.Status.Plugins).To(HaveKey(storage.PluginName))
					g.Expect(updated.Status.Plugins[storage.PluginName].Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
					g.Expect(updated.Status.Plugins[storage.PluginName].Conditions).To(BeEmpty())
				}, timeout, interval).Should(Succeed())

				pvcs := &v1.PersistentVolumeClaimList{}
				Expect(k8sClient.List(managerCtx, pvcs, client.InNamespace(storageNamespace))).To(Succeed())
				Expect(pvcs.Items).To(BeEmpty())
			},
			Entry("when Instructions is nil", nil),
			Entry("when Instructions.Storage is nil", &dromosv1alpha1.Instructions{}),
			Entry("when Instructions.Storage.NewPvcs is empty", &dromosv1alpha1.Instructions{Storage: &dromosv1alpha1.StorageConfig{}}),
		)
	})

	Context("WorkloadRunner with PVC validation", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		DescribeTable("should reject WorkloadRunner",
			func(mutate func(*v1.PersistentVolumeClaim), expectedSubstring string) {
				pvc := newDefaultPvc("data")
				mutate(&pvc)
				runner := newStorageRunner(storageNamespace, nil, pvc)

				err := k8sClient.Create(managerCtx, runner)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedSubstring))
			},
			Entry("empty PVC name", func(p *v1.PersistentVolumeClaim) { p.Name = "" }, "name is required"),
			Entry("PVC namespace != WorkloadRunner.namespace", func(p *v1.PersistentVolumeClaim) { p.Namespace = "some-other-ns" }, "namespace must match WorkloadRunner"),
			Entry("empty accessModes", func(p *v1.PersistentVolumeClaim) { p.Spec.AccessModes = nil }, "at least one access mode is required"),
			Entry("missing storage request", func(p *v1.PersistentVolumeClaim) { p.Spec.Resources.Requests = nil }, "storage request is required"),
		)

		It("should reject WorkloadRunner with duplicate PVC names", func() {
			runner := newStorageRunner(storageNamespace, nil, newDefaultPvc("dup"), newDefaultPvc("dup"))

			err := k8sClient.Create(managerCtx, runner)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Duplicate value"))
		})

		It("should accept WorkloadRunner with well-formed PVC spec", func() {
			runner := newStorageRunner(storageNamespace, nil, newDefaultPvc("ok"))
			Expect(k8sClient.Create(managerCtx, runner)).To(Succeed())
		})
	})

	Context("WorkloadRunner with PVC creation", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		It("creates a single PVC with ownerRef to the WorkloadRunner", func() {
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			pvcName := "pvc-" + rand.String(4)
			runner := createStorageRunner(managerCtx, k8sClient, storageNamespace, newDefaultPvc(pvcName))

			created := &v1.PersistentVolumeClaim{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: storageNamespace}, created)).To(Succeed())
				g.Expect(created.OwnerReferences).To(HaveLen(1))
				g.Expect(created.OwnerReferences[0].UID).To(Equal(runner.UID))
				g.Expect(created.OwnerReferences[0].Kind).To(Equal("WorkloadRunner"))
				g.Expect(created.OwnerReferences[0].Controller).NotTo(BeNil())
				g.Expect(*created.OwnerReferences[0].Controller).To(BeTrue())
				g.Expect(metav1.IsControlledBy(created, runner)).To(BeTrue())
				wr := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
				g.Expect(wr.Status.Plugins[storage.PluginName].Phase).ToNot(And(Equal(dromosv1alpha1.PluginPhaseFailed), Equal(dromosv1alpha1.PluginPhaseDisabled)))
			}, timeout, interval).Should(Succeed())
		})

		It("creates multiple PVCs in one runner", func() {
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			name1, name2 := "pvc-"+rand.String(4), "pvc-"+rand.String(4)
			pvc1, pvc2 := newDefaultPvc(name1), newDefaultPvc(name2)
			_ = createStorageRunner(managerCtx, k8sClient, storageNamespace, pvc1, pvc2)
			Eventually(func(g Gomega) {
				for _, n := range []string{name1, name2} {
					pvc := &v1.PersistentVolumeClaim{}
					g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: n, Namespace: storageNamespace}, pvc)).To(Succeed())
				}
			}, timeout, interval).Should(Succeed())
		})

		It("defaults PVC namespace to runner namespace when empty", func() {
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			pvc := newDefaultPvc("pvc-" + rand.String(4))
			pvc.Namespace = ""
			runner := createStorageRunner(managerCtx, k8sClient, storageNamespace, pvc)

			created := &v1.PersistentVolumeClaim{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: pvc.Name, Namespace: runner.Namespace}, created)).To(Succeed())
				g.Expect(created.Namespace).To(Equal(runner.Namespace))
			}, timeout, interval).Should(Succeed())
		})

		It("re-applies idempotently when PVC already owned by same runner", func() {
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			pvcName := "pvc-" + rand.String(4)
			runner := createStorageRunner(managerCtx, k8sClient, storageNamespace, newDefaultPvc(pvcName))

			created := &v1.PersistentVolumeClaim{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: storageNamespace}, created)
			}, timeout, interval).Should(Succeed())
			firstUID := created.UID

			// Synchronously invoke another reconcile N times
			for range 5 {
				result, err := storageController.Reconcile(
					managerCtx,
					ctrl.Request{
						NamespacedName: types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace},
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(BeZero())
			}

			// PVC must not be recreated; plugin must not be Failed.
			pvc := &v1.PersistentVolumeClaim{}
			Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: storageNamespace}, pvc)).To(Succeed())
			Expect(pvc.UID).To(Equal(firstUID))

			wr := &dromosv1alpha1.WorkloadRunner{}
			Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
			if p, ok := wr.Status.Plugins[storage.PluginName]; ok {
				Expect(p.Phase).NotTo(Equal(dromosv1alpha1.PluginPhaseFailed))
			}
		})
	})

	Context("WorkloadRunner with PVC ownership conflict", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		DescribeTable("emits critical issue when PVC exists not controlled by this runner",
			func(ownerRefs []metav1.OwnerReference) {
				// Pre-create a PVC with the conflicting ownership state.
				pvcName := "pvc-" + rand.String(4)
				existing := newDefaultPvc(pvcName)
				existing.Namespace = storageNamespace
				existing.OwnerReferences = ownerRefs
				Expect(k8sClient.Create(managerCtx, &existing)).To(Succeed())

				// Create a WR claiming the same PVC name.
				_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				runner := createStorageRunner(managerCtx, k8sClient, storageNamespace, newDefaultPvc(pvcName))

				// Plugin should hit the ownership-conflict the controller sets the WR's Plugin status AND overall Status.Phase to Failed.
				Eventually(func(g Gomega) {
					wr := &dromosv1alpha1.WorkloadRunner{}
					g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
					g.Expect(wr.Status.Plugins).To(HaveKey(storage.PluginName))
					g.Expect(wr.Status.Plugins[storage.PluginName].Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
					g.Expect(wr.Status.Plugins[storage.PluginName].Issues).NotTo(BeEmpty())
					g.Expect(wr.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseFailed))
				}, timeout, interval).Should(Succeed())
			},
			Entry("PVC has no ownerRef (unowned)", []metav1.OwnerReference(nil)),
			Entry("PVC owned by a different WorkloadRunner", []metav1.OwnerReference{{
				APIVersion: "run.ai/v1alpha1",
				Kind:       "WorkloadRunner",
				Name:       "other-runner",
				UID:        types.UID("fake-uid-other-wr"),
				Controller: new(true),
			}}),
		)
	})

	Context("WorkloadRunner plugin status phase tree", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		DescribeTable("derives plugin phase from underlying PVC status.phase",
			func(targetPvcPhase v1.PersistentVolumeClaimPhase, expectedPluginPhase dromosv1alpha1.PluginPhase) {
				_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				pvcName := "pvc-" + rand.String(4)
				runner := createStorageRunner(managerCtx, k8sClient, storageNamespace, newDefaultPvc(pvcName))

				live := &v1.PersistentVolumeClaim{}
				Eventually(func() error {
					return k8sClient.Get(managerCtx,
						types.NamespacedName{Name: pvcName, Namespace: storageNamespace}, live)
				}, timeout, interval).Should(Succeed())

				// Pending is the envtest default after Create, so no /status write needed.
				if targetPvcPhase != v1.ClaimPending {
					live.Status.Phase = targetPvcPhase
					Expect(k8sClient.Status().Update(managerCtx, live)).To(Succeed())
				}

				Eventually(func(g Gomega) {
					wr := &dromosv1alpha1.WorkloadRunner{}
					g.Expect(k8sClient.Get(managerCtx,
						types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
					g.Expect(wr.Status.Plugins).To(HaveKey(storage.PluginName))
					g.Expect(wr.Status.Plugins[storage.PluginName].Phase).To(Equal(expectedPluginPhase))
				}, timeout, interval).Should(Succeed())
			},
			Entry("Pvc: Pending, Plugin: Initializing", v1.ClaimPending, dromosv1alpha1.PluginPhaseInitializing),
			Entry("Pvc: Bound, Plugin: Ready", v1.ClaimBound, dromosv1alpha1.PluginPhaseReady),
			Entry("Pvc: Lost, Plugin: Failed", v1.ClaimLost, dromosv1alpha1.PluginPhaseFailed),
		)
	})
	Context("PVC cleanup on terminal phase", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		DescribeTable("deletes owned PVCs only when cleanup is enabled on a terminal phase",
			func(cleanupOnCompletion *bool, conditionType string, expectedPhase dromosv1alpha1.WorkloadRunnerPhase, expectDeleted bool) {
				pvcName := "pvc-" + rand.String(4)
				_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
				spec := newStorageRunner(storageNamespace, cleanupOnCompletion, newDefaultPvc(pvcName))
				runner := createResource(managerCtx, k8sClient, spec)
				workloadName := runner.ObjectMeta.Name

				Eventually(func() error {
					return k8sClient.Get(managerCtx, types.NamespacedName{Name: pvcName, Namespace: storageNamespace}, &v1.PersistentVolumeClaim{})
				}, timeout, interval).Should(Succeed())

				updateJobSetStatus(managerCtx, k8sClient, workloadName, storageNamespace, []any{makeCondition(conditionType, "True")})
				Eventually(func(g Gomega) {
					wr := &dromosv1alpha1.WorkloadRunner{}
					g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
					g.Expect(wr.Status.Phase).To(Equal(expectedPhase))
					g.Expect(wr.Status.ResourcePhase).To(Equal(expectedPhase))
				}, timeout, interval).Should(Succeed())

				if expectDeleted {
					Eventually(pvcDeletionIssued(managerCtx, k8sClient, storageNamespace, pvcName), timeout, interval).Should(BeTrue(), "expected PVC %q to be deleted", pvcName)
				} else {
					Consistently(pvcDeletionIssued(managerCtx, k8sClient, storageNamespace, pvcName), "1s", interval).Should(BeFalse(), "expected PVC %q to be untouched", pvcName)
				}
			},
			Entry("cleanup not set,workload Completed, not deleted", nil, "Completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, false),
			Entry("cleanup not set,workload Failed, not deleted", nil, "Failed", dromosv1alpha1.WorkloadRunnerPhaseFailed, false),
			Entry("cleanup=false, Completed, not deleted", new(false), "Completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, false),
			Entry("cleanup=true, Completed, deleted", new(true), "Completed", dromosv1alpha1.WorkloadRunnerPhaseCompleted, true),
			Entry("cleanup=true, Failed, deleted", new(true), "Failed", dromosv1alpha1.WorkloadRunnerPhaseFailed, true),
		)
	})

	Context("PVC creation skipped on terminal phase", func() {
		AfterEach(func() {
			deleteAllTestResources(managerCtx, k8sClient)
			deleteAllForType(managerCtx, k8sClient, &v1.PersistentVolumeClaimList{})
		})

		It("does not (re-)create a PVC once the runner is already terminal with cleanup enabled", func() {
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			pvcName := "pvc-" + rand.String(4)
			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			spec := newStorageRunner(storageNamespace, new(true), newDefaultPvc(pvcName))
			runner := createResource(managerCtx, k8sClient, spec)
			workloadName := runner.ObjectMeta.Name

			// Reach Completed once, which deletes the PVC created on the first reconcile.
			updateJobSetStatus(managerCtx, k8sClient, workloadName, storageNamespace, []any{makeCondition("Completed", "True")})
			Eventually(func(g Gomega) {
				wr := &dromosv1alpha1.WorkloadRunner{}
				g.Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, wr)).To(Succeed())
				g.Expect(wr.Status.Phase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseCompleted))
				g.Expect(wr.Status.ResourcePhase).To(Equal(dromosv1alpha1.WorkloadRunnerPhaseCompleted))
			}, timeout, interval).Should(Succeed())
			Eventually(pvcDeletionIssued(managerCtx, k8sClient, storageNamespace, pvcName), timeout, interval).Should(BeTrue(), "expected PVC %q to be deleted", pvcName)

			// The PVC deletion event (Owns watch) triggers further reconciles of the
			// now-terminal runner; the PVC must stay deleted and not be recreated.
			Consistently(pvcDeletionIssued(managerCtx, k8sClient, storageNamespace, pvcName), "1s", interval).Should(BeTrue(), "expected PVC %q to stay deleted/terminating, not be recreated", pvcName)
		})
	})
})

// pvcDeletionIssued reports whether the PVC is gone or has a deletion timestamp.
func pvcDeletionIssued(ctx context.Context, k8sClient client.Client, namespace, name string) func() bool {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	return func() bool {
		pvc := &v1.PersistentVolumeClaim{}
		err := k8sClient.Get(ctx, key, pvc)
		return apierrors.IsNotFound(err) || (err == nil && pvc.DeletionTimestamp != nil)
	}
}

func newDefaultPvc(name string) v1.PersistentVolumeClaim {
	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Gi")},
			},
			StorageClassName: new("standard"),
		},
	}
}

func createStorageRunner(ctx context.Context, k8sClient client.Client, namespace string, pvcs ...v1.PersistentVolumeClaim) *dromosv1alpha1.WorkloadRunner {
	return createResource(ctx, k8sClient, newStorageRunner(namespace, nil, pvcs...))
}

func newStorageRunner(namespace string, cleanupOnCompletion *bool, pvcs ...v1.PersistentVolumeClaim) *dromosv1alpha1.WorkloadRunner {
	suffix := rand.String(4)
	name := "test-runner-" + suffix
	jobset := test.NewJobSetBuilder(name, namespace).WithDefault().Build()
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: marshalRawExtension(jobset),
			Instructions: &dromosv1alpha1.Instructions{
				Storage: &dromosv1alpha1.StorageConfig{
					NewPvcs:             pvcs,
					CleanupOnCompletion: cleanupOnCompletion,
				},
			},
		},
	}
}
