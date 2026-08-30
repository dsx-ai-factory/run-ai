package deletion_test

import (
	"context"
	"time"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/reconcilers"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	v1 "k8s.io/api/core/v1"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Project Finalization", func() {

	var (
		reconciler *ProjectReconciler
		client     cli.Client
		project    kaiv1alpha1.Project
		namespace  v1.Namespace
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()

		project.Finalizers = append(project.Finalizers, config.FinalizerName())
		client = fake.NewClientBuilder().WithScheme(scheme).Build()
		reconciler = NewProjectReconciler(client, client, nil, nil, config.NewProjectReconcilerConfig())
		// We're not testing the deletion handlers in this test
		SetUnexportedField(reconciler, "deletionHandlers", []ProjectResourceDeletionHandler{})

		_ = client.Create(context.TODO(), &namespace)
	})

	It("Adds the controller as finalizer", func() {
		When("Adds the controller as finalizer on the project", func() {

			// Given
			Expect(client.Create(context.TODO(), &project)).To(Succeed())

			// When
			_ = reconciler.AddControllerAsFinalizerIfNeeded(context.TODO(), &project)

			// Then
			var actualProject kaiv1alpha1.Project
			Expect(client.Get(context.TODO(), cli.ObjectKey{Name: project.Name}, &actualProject)).To(Succeed())
			Expect(actualProject.Finalizers).To(HaveLen(1))
			Expect(actualProject.Finalizers[0]).To(Equal(config.FinalizerName()))
		})

	})

	It("Doesn't add the controller as finalizer", func() {
		When("the finalizer entry already exists", func() {

			// Given
			project.Finalizers = []string{config.FinalizerName()}
			Expect(client.Create(context.TODO(), &project)).To(Succeed())

			// When
			_ = reconciler.AddControllerAsFinalizerIfNeeded(context.TODO(), &project)

			// Then
			var actualProject kaiv1alpha1.Project
			Expect(client.Get(context.TODO(), cli.ObjectKey{Name: project.Name}, &actualProject)).To(Succeed())
			Expect(actualProject.Finalizers).To(HaveLen(1))
			Expect(actualProject.Finalizers[0]).To(Equal(config.FinalizerName()))
		})

	})

	Context("using test deletion handler", func() {

		var testDeletionHandler TestDeletionHandler

		JustBeforeEach(func() {
			testDeletionHandler = TestDeletionHandler{}

			// To test if finalization is called or not we'll add a deletion handler and test if its logic was executed
			SetUnexportedField(reconciler, "deletionHandlers", []ProjectResourceDeletionHandler{&testDeletionHandler})
		})

		When("the finalizer entry doesn't exist on the project", func() {
			It("it doesn't call the deletion handlers", func() {

				// Given
				project.Finalizers = []string{}
				Expect(client.Create(context.TODO(), &project)).To(Succeed())

				// When
				res, err := reconciler.Finalize(context.Background(), &project)

				// Then
				Expect(res).To(BeZero())
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeletionHandler.HandlerCalled).To(BeFalse())
			})
		})

		When("the project is manually overridden", func() {
			It("it doesn't call the deletion handlers", func() {

				// Given
				project.SetLabels(map[string]string{RunaiResourceManualOverrideLabel: "true"})
				Expect(client.Create(context.TODO(), &project)).To(Succeed())

				// When
				res, err := reconciler.Finalize(context.Background(), &project)

				// Then
				Expect(res).To(BeZero())
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeletionHandler.HandlerCalled).To(BeFalse())
			})
		})

		When("the finalizer entry exists", func() {
			It("it calls the deletion handlers", func() {

				// Given
				project.Finalizers = []string{config.FinalizerName()}
				Expect(client.Create(context.TODO(), &project)).To(Succeed())

				// When
				res, err := reconciler.Finalize(context.Background(), &project)

				// Then
				Expect(res).To(BeZero())
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeletionHandler.HandlerCalled).To(BeTrue())
				Expect(project.Finalizers).To(HaveLen(0)) // Finalizer entry should be deleted if finalization is successful
			})
		})

		When("the namespace does not exist", func() {
			It("it does not call the deletion handlers", func() {
				// Given
				project.Finalizers = []string{config.FinalizerName()}
				Expect(client.Create(context.TODO(), &project)).To(Succeed())

				ns := v1.Namespace{}
				Eventually(func() error {
					return client.Get(context.Background(), cli.ObjectKey{Name: namespace.Name}, &ns)
				}, 1*time.Second).Should(Succeed())
				Expect(client.Delete(context.TODO(), &ns)).To(Succeed())

				// When
				res, err := reconciler.Finalize(context.Background(), &project)

				// Then
				Expect(res).To(BeZero())
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeletionHandler.HandlerCalled).To(BeTrue())
				Expect(project.Finalizers).To(HaveLen(0)) // Finalizer entry should be deleted if finalization is successful
			})
		})
	})
})
