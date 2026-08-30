package test

import (
	"context"
	"testing"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/consts"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/providers"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	toolscache "k8s.io/client-go/tools/cache"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/handler"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApplicationHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Handler Suite")
}

var _ = Describe("StartApplicationHandler", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
		k8sCli client.Client
		cache  *MockCache
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(v2.AddToScheme(scheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())
		k8sCli = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&runv1alpha1.Application{}).Build()
		cache = &MockCache{Cli: k8sCli, WaitForSyncResult: true}
	})

	Describe("handler start", func() {
		It("registers a FilteringResourceEventHandler on all informers and waits for sync", func() {
			Expect(func() { handler.StartApplicationHandler(ctx, k8sCli, cache) }).NotTo(Panic())
			Expect(cache.GetInformerCalled).To(BeTrue())
			Expect(cache.Informers).To(HaveLen(1))
			Expect(cache.Informers[0].Handler).NotTo(BeNil()) // Helm informer
		})
	})

	Describe("secret filter", func() {
		It("namespace not runai- no label", func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			handler.StartApplicationHandler(ctx, k8sCli, cache)

			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			f, ok := cache.Informers[0].Handler.(toolscache.FilteringResourceEventHandler)
			Expect(ok).To(BeTrue())
			Expect(f.FilterFunc(secret)).To(BeFalse())
		})

		It("namespace not runai- no annotation", func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "true"},
				Annotations: map[string]string{},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			handler.StartApplicationHandler(ctx, k8sCli, cache)

			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			f, ok := cache.Informers[0].Handler.(toolscache.FilteringResourceEventHandler)
			Expect(ok).To(BeTrue())
			Expect(f.FilterFunc(secret)).To(BeFalse())
		})

		It("secret not helm", func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "true"},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			handler.StartApplicationHandler(ctx, k8sCli, cache)

			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
			}, Type: corev1.SecretTypeOpaque}

			f, ok := cache.Informers[0].Handler.(toolscache.FilteringResourceEventHandler)
			Expect(ok).To(BeTrue())
			Expect(f.FilterFunc(secret)).To(BeFalse())
		})
	})

	Describe("application update", func() {
		BeforeEach(func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "true"},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			// Pre-create Application
			app := &runv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Name: "my-app", Namespace: "ns1"}}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())
			// Seed resources with label/annotation
			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
				Name: "ew-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "cm-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name: "sec-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())

			handler.StartApplicationHandler(ctx, k8sCli, cache)
		})

		It("if application exists, updates status with different phase and resources (no create)", func() {
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			cache.Informers[0].Handler.OnAdd(secret, false)

			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal("deployed"))

			names := map[string]bool{}
			for _, r := range updated.Status.Resources {
				names[r.Name] = true
			}
			Expect(names["ew-1"]).To(BeTrue())
			Expect(names["pvc-1"]).To(BeTrue())
			Expect(names["cm-1"]).To(BeTrue())
			Expect(names["sec-1"]).To(BeTrue())

			// Validate ApplicationIdLabel and ApplicationNameLabel are set on the ExternalWorkload
			updatedEW := &runv1alpha1.ExternalWorkload{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "ew-1", Namespace: "ns1"}, updatedEW)).To(Succeed())
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationIdLabel), "ExternalWorkload should have ApplicationIdLabel")
			Expect(updatedEW.Labels[constants.ApplicationIdLabel]).To(Equal(string(updated.UID)), "ApplicationIdLabel should match Application UID")
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationNameLabel), "ExternalWorkload should have ApplicationNameLabel")
			Expect(updatedEW.Labels[constants.ApplicationNameLabel]).To(Equal(updated.Name), "ApplicationNameLabel should match Application Name")
		})

		It("defaults application phase to unknown when label missing", func() {
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app"},
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			cache.Informers[0].Handler.OnAdd(secret, false)

			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal(consts.UnknownApplicationPhase))
		})
	})

	Describe("application resources update", func() {
		BeforeEach(func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "true"},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			// Pre-create Application
			app := &runv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Name: "my-app", Namespace: "ns1"}, Status: runv1alpha1.ApplicationStatus{
				K8sPhase: "unknown",
				Resources: []runv1alpha1.ApplicationResource{{
					Name: "ew-deleted",
					Kind: "",
					UID:  "9052827a-a77d-4fd5-8575-845eaf709822",
				}},
			}}
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			handler.StartApplicationHandler(ctx, k8sCli, cache)
		})

		It("if resources deleted they are removed from status", func() {
			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal("unknown"))

			names := map[string]bool{}
			for _, r := range updated.Status.Resources {
				names[r.Name] = true
			}
			Expect(len(names)).To(Equal(1))
			Expect(names["ew-1"]).To(BeFalse())
			Expect(names["pvc-1"]).To(BeFalse())
			Expect(names["cm-1"]).To(BeFalse())
			Expect(names["sec-1"]).To(BeFalse())
			Expect(names["ew-deleted"]).To(BeTrue())

			// Seed resources with label/annotation
			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
				Name: "ew-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "cm-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name: "sec-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())

			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			cache.Informers[0].Handler.OnAdd(secret, false)

			updated = &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal("deployed"))

			names = map[string]bool{}
			for _, r := range updated.Status.Resources {
				names[r.Name] = true
			}
			Expect(len(names)).To(Equal(4))
			Expect(names["ew-1"]).To(BeTrue())
			Expect(names["pvc-1"]).To(BeTrue())
			Expect(names["cm-1"]).To(BeTrue())
			Expect(names["sec-1"]).To(BeTrue())
			Expect(names["ew-deleted"]).To(BeFalse())

			// Validate ApplicationIdLabel and ApplicationNameLabel are set on the ExternalWorkload
			updatedEW := &runv1alpha1.ExternalWorkload{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "ew-1", Namespace: "ns1"}, updatedEW)).To(Succeed())
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationIdLabel), "ExternalWorkload should have ApplicationIdLabel")
			Expect(updatedEW.Labels[constants.ApplicationIdLabel]).To(Equal(string(updated.UID)), "ApplicationIdLabel should match Application UID")
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationNameLabel), "ExternalWorkload should have ApplicationNameLabel")
			Expect(updatedEW.Labels[constants.ApplicationNameLabel]).To(Equal(updated.Name), "ApplicationNameLabel should match Application Name")
		})
	})

	Describe("application create", func() {
		BeforeEach(func() {
			// Seed namespace with required label/annotation
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "proj-1"},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			}}
			Expect(k8sCli.Create(ctx, ns)).To(Succeed())
			// Seed resources with label/annotation
			Expect(k8sCli.Create(ctx, &v2.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "proj-1",
					Labels: map[string]string{v2.ProjectIdLabel: "11", v2alpha2.DepartmentIdLabel: "1"},
				},
				Spec: v2.ProjectSpec{Department: "dept-name-1"},
			})).To(Succeed())
			// kai.resources twin used by the suite
			// (the resolver reads whichever the flag selects). Spec.Parent replaces Spec.Department.
			Expect(k8sCli.Create(ctx, &kaiv1alpha1.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "proj-1",
					Labels: map[string]string{v2.ProjectIdLabel: "11", v2alpha2.DepartmentIdLabel: "1"},
				},
				Spec: kaiv1alpha1.ProjectSpec{Parent: "dept-name-1"},
			})).To(Succeed())
			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
				Name: "ew-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "cm-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name: "sec-1", Namespace: "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{constants.ResourceReleaseAnnotationName: "my-app"},
			}})).To(Succeed())

			// Inject mock resolver for testing
			handler.GetHelmSourceProvider().SetObjectResolver(&MockObjectResolver{})
			handler.StartApplicationHandler(ctx, k8sCli, cache)
		})

		It("if application does not exists, create + updates status with resources", func() {
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
				Name:      "sh.helm.release.v1.my-app.v1",
				UID:       "9052827a-a77d-4fd5-8575-845eaf709821",
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			cache.Informers[0].Handler.OnAdd(secret, false)

			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal("deployed"))
			Expect(len(updated.OwnerReferences)).To(Equal(1))
			Expect(updated.OwnerReferences[0].Kind).To(Equal("Secret"))
			Expect(updated.OwnerReferences[0].UID).To(Equal(secret.UID))
			Expect(updated.OwnerReferences[0].Name).To(Equal(secret.Name))
			Expect(updated.Labels[constants.ProjectNameLabel]).To(Equal("proj-1"))
			Expect(updated.Labels[constants.ProjectIdLabel]).To(Equal("11"))
			Expect(updated.Labels[constants.DepartmentIdLabel]).To(Equal("1"))
			Expect(updated.Labels[constants.DepartmentNameLabel]).To(Equal("dept-name-1"))

			names := map[string]bool{}
			for _, r := range updated.Status.Resources {
				names[r.Name] = true
			}
			Expect(names["ew-1"]).To(BeTrue())
			Expect(names["pvc-1"]).To(BeTrue())
			Expect(names["cm-1"]).To(BeTrue())
			Expect(names["sec-1"]).To(BeTrue())

			// Validate ApplicationIdLabel and ApplicationNameLabel are set on the ExternalWorkload
			updatedEW := &runv1alpha1.ExternalWorkload{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "ew-1", Namespace: "ns1"}, updatedEW)).To(Succeed())
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationIdLabel), "ExternalWorkload should have ApplicationIdLabel")
			Expect(updatedEW.Labels[constants.ApplicationIdLabel]).To(Equal(string(updated.UID)), "ApplicationIdLabel should match Application UID")
			Expect(updatedEW.Labels).To(HaveKey(constants.ApplicationNameLabel), "ExternalWorkload should have ApplicationNameLabel")
			Expect(updatedEW.Labels[constants.ApplicationNameLabel]).To(Equal(updated.Name), "ApplicationNameLabel should match Application Name")
		})

		It("if application already exists without the department-name label, backfills it (cluster upgrade)", func() {
			// Simulate an Application created by an older cluster version: it has the
			// project/department-id labels but not the department-name label.
			Expect(k8sCli.Create(ctx, &runv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{
				Name:      "my-app",
				Namespace: "ns1",
				Labels: map[string]string{
					constants.ProjectNameLabel:  "proj-1",
					constants.ProjectIdLabel:    "11",
					constants.DepartmentIdLabel: "1",
				},
			}})).To(Succeed())

			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns1",
				Labels:    map[string]string{consts.SecretNameLabel: "my-app", consts.SecretPhaseLabel: "deployed"},
				Name:      "sh.helm.release.v1.my-app.v1",
				UID:       "9052827a-a77d-4fd5-8575-845eaf709821",
			}, Type: corev1.SecretType(providers.HelmSecretType)}

			cache.Informers[0].Handler.OnAdd(secret, false)

			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Labels[constants.DepartmentNameLabel]).To(Equal("dept-name-1"))
			// Existing labels are preserved.
			Expect(updated.Labels[constants.DepartmentIdLabel]).To(Equal("1"))
			Expect(updated.Labels[constants.ProjectIdLabel]).To(Equal("11"))
		})
	})
})
