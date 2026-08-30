package reconcilers_test

import (
	"context"
	"time"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/reconcilers"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// +kubebuilder:scaffold:imports
)

const (
	testHandlerErrorStr = "fake error"
)

var (
	projectReconcileRequest = ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "",
			Name:      TestProject.Name,
		},
	}

	now = metav1.NewTime(time.Now())
)

var _ = Describe("Project Reconciler Tests", func() {

	var (
		reconciler               *ProjectReconciler
		k8sClient                client.Client
		projectEvents            chan event.GenericEvent
		project, modifiedProject kaiv1alpha1.Project
		namespace                v1.Namespace
		newNamespaceName         string
		queue                    kaiv2.Queue
		dep                      kaiv1alpha1.Department
		projectConfig            *config.ProjectReconcilerConfig
	)

	BeforeEach(func() {
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiv1alpha1.Project{}).
			WithStatusSubresource(&kaiv2.Queue{}).
			Build()
		projectEvents = make(chan event.GenericEvent)
		projectConfig = config.NewProjectReconcilerConfig()
		projectConfig.CreateNamespaces = true
		projectConfig.CreateRoleBindings = true
		projectConfig.NodePoolLabelKey = NodePoolLabelKey
		reconciler = NewProjectReconciler(k8sClient, k8sClient, scheme, projectEvents, projectConfig)
		project = *TestProject.DeepCopy()
		modifiedProject = *KaiModifiedTestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()
		queue = *TestQueue.DeepCopy()
		dep = *KaiTestDepartment1.DeepCopy()
		newNamespaceName = handlers.DefaultProjectNamespaceName(&project)

		// will create the parent queues that project Reconcile is expecting
		departmentHandler := handlers.NewDepartmentHandler(k8sClient)
		err := departmentHandler.Handle(context.Background(), &dep)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Reconciles a new Project", func() {

		// Given
		sanity := kaiv1alpha1.Project{}
		Expect(k8sClient.Create(context.TODO(), &project)).To(Succeed())
		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).To(Succeed())
		Expect(sanity.Name).To(Equal(project.Name))
		Expect(sanity.Finalizers).To(HaveLen(0))

		// When
		result, err := reconciler.Reconcile(context.TODO(), projectReconcileRequest)

		// Then
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		By("Adding the reconciler to the finalizers list", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Name).To(Equal(project.Name))
			Expect(actualProject.Finalizers).To(HaveLen(1))
			Expect(actualProject.Finalizers[0]).To(Equal(config.FinalizerName()))
		})

		By("Calling all reconcilers", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Name).To(Equal(project.Name))

			for _, queue := range actualProject.Spec.Queues {
				actualQueue := kaiv2.Queue{}
				Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: queue.Name}, &actualQueue)).To(Succeed())
				Expect(actualQueue.Name).To(Equal(queue.Name))
				Expect(actualQueue.OwnerReferences[0].UID).To(Equal(actualProject.UID))
				if queue.Nodepool != config.Get().DefaultNodepoolName {
					Expect(actualQueue.Labels[NodePoolLabelKey]).To(Equal(queue.Nodepool))
				} else {
					Expect(actualQueue.Labels).ToNot(HaveKey(NodePoolLabelKey))
				}
			}

			actualNamespace := v1.Namespace{}

			Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: newNamespaceName}, &actualNamespace)).To(Succeed())
			Expect(actualNamespace.Name).To(Equal(newNamespaceName))
			Expect(actualNamespace.OwnerReferences).To(BeEmpty())
			Expect(actualProject.Status.Namespace).To(Equal(newNamespaceName))
		})

		By("Correct Project Status", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Status.Namespace).To(Equal(newNamespaceName))
			Expect(actualProject.Status.Phase).To(Equal(kaiv1alpha1.Ready))
			Expect(actualProject.Status.Message).To(Equal(""))
			Expect(len(actualProject.Status.Conditions)).ToNot(Equal(0))
		})

		By("Correct Queue-aggregated Project Status", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())

			// update queue statuses to validate project queue-aggregated status
			for _, queue := range actualProject.Spec.Queues {
				actualQueue := kaiv2.Queue{}
				Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: queue.Name}, &actualQueue)).To(Succeed())
				actualQueue.Status = *(QueueStatusForTests.DeepCopy())
				Expect(k8sClient.Status().Update(context.TODO(), &actualQueue)).To(Succeed())
			}

			// now Reconcile again to update the status
			result, err = reconciler.Reconcile(context.TODO(), projectReconcileRequest)
			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())
			actualProject = kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())

			Expect(actualProject.Status.Namespace).To(Equal(newNamespaceName))
			Expect(actualProject.Status.Phase).To(Equal(kaiv1alpha1.Ready))
			Expect(actualProject.Status.Message).To(Equal(""))
			Expect(len(actualProject.Status.Conditions)).ToNot(Equal(0))

			// validate quota statuses correct
			Expect(len(actualProject.Status.NodePoolsQuotaStatuses)).To(Equal(len(actualProject.Spec.Queues)))
			queueStatusesMap := map[string]kaiv2.QueueStatus{}
			for _, quotaStatus := range actualProject.Status.NodePoolsQuotaStatuses {
				validateQueueStatusEqual(&quotaStatus.QueueStatus, &QueueStatusForTests)

				queueStatusesMap[quotaStatus.NodePoolName] = quotaStatus.QueueStatus
			}
			for _, queueSpec := range actualProject.Spec.Queues {
				_, found := queueStatusesMap[queueSpec.Nodepool]
				Expect(found).To(BeTrue(), "expected to find queue status for nodepool <%s>",
					queueSpec.Nodepool)
			}

			// remove one of the nodepool-queues from project spec
			actualProject.Spec.Queues = []kaiv1alpha1.QueueConfig{actualProject.Spec.Queues[0]}
			Expect(k8sClient.Update(context.TODO(), &actualProject)).To(Succeed())

			// now Reconcile again to update the status
			result, err = reconciler.Reconcile(context.TODO(), projectReconcileRequest)
			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())
			actualProject = kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())

			// validate quota status correct
			Expect(len(actualProject.Status.NodePoolsQuotaStatuses)).To(Equal(len(actualProject.Spec.Queues)))
			Expect(len(actualProject.Status.NodePoolsQuotaStatuses)).To(Equal(1))
			validateQueueStatusEqual(&(actualProject.Status.NodePoolsQuotaStatuses[0].QueueStatus), &QueueStatusForTests)
			Expect(actualProject.Status.NodePoolsQuotaStatuses[0].NodePoolName).To(Equal(actualProject.Spec.Queues[0].Nodepool))
		})
	})

	It("Clears Project quota status when Queue quota status is emptied", func() {
		// This test verifies that when queue allocation/requested resources are cleared,
		// the project status correctly reflects the empty quota (not retaining old values)

		// Given - create project and reconcile
		Expect(k8sClient.Create(context.TODO(), &project)).To(Succeed())
		result, err := reconciler.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		// Step 1: Update queue statuses with allocation data
		actualProject := kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())

		for _, queueSpec := range actualProject.Spec.Queues {
			actualQueue := kaiv2.Queue{}
			Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: queueSpec.Name}, &actualQueue)).To(Succeed())
			actualQueue.Status = *(QueueStatusForTests.DeepCopy())
			Expect(k8sClient.Status().Update(context.TODO(), &actualQueue)).To(Succeed())
		}

		// Reconcile to update project status with quota data
		result, err = reconciler.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		// Verify project has quota data
		actualProject = kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
		Expect(len(actualProject.Status.NodePoolsQuotaStatuses)).To(BeNumerically(">", 0))

		// Verify QuotaStatus has aggregated values
		Expect(actualProject.Status.QuotaStatus.Allocated).ToNot(BeEmpty(),
			"Project QuotaStatus.Allocated should have values after queue status update")
		Expect(actualProject.Status.QuotaStatus.AllocatedNonPreemptible).ToNot(BeEmpty(),
			"Project QuotaStatus.AllocatedNonPreemptible should have values after queue status update")
		Expect(actualProject.Status.QuotaStatus.Requested).ToNot(BeEmpty(),
			"Project QuotaStatus.Requested should have values after queue status update")

		// Verify the actual aggregated values (sum of all queues)
		// QueueStatusForTests has: Allocated["nvidia.com/gpu"] = 1000m, Requested = 2000m, AllocatedNonPreemptible = 500m
		// With 2 queues, expected sums are: Allocated = 2000m, Requested = 4000m, AllocatedNonPreemptible = 1000m
		Expect(len(actualProject.Spec.Queues)).To(Equal(2), "Test expects 2 queues for validation")

		actualGpuAllocated := actualProject.Status.QuotaStatus.Allocated["nvidia.com/gpu"]
		Expect(actualGpuAllocated.MilliValue()).To(Equal(int64(2000)),
			"Project QuotaStatus.Allocated[nvidia.com/gpu] should be 2000m (sum of 2 queues x 1000m)")

		actualGpuRequested := actualProject.Status.QuotaStatus.Requested["nvidia.com/gpu"]
		Expect(actualGpuRequested.MilliValue()).To(Equal(int64(4000)),
			"Project QuotaStatus.Requested[nvidia.com/gpu] should be 4000m (sum of 2 queues x 2000m)")

		actualGpuAllocatedNonPreemptible := actualProject.Status.QuotaStatus.AllocatedNonPreemptible["nvidia.com/gpu"]
		Expect(actualGpuAllocatedNonPreemptible.MilliValue()).To(Equal(int64(1000)),
			"Project QuotaStatus.AllocatedNonPreemptible[nvidia.com/gpu] should be 1000m (sum of 2 queues x 500m)")

		// Step 2: Clear queue statuses (empty allocation/requested)
		for _, queueSpec := range actualProject.Spec.Queues {
			actualQueue := kaiv2.Queue{}
			Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: queueSpec.Name}, &actualQueue)).To(Succeed())
			actualQueue.Status = kaiv2.QueueStatus{
				Allocated:               v1.ResourceList{},
				AllocatedNonPreemptible: v1.ResourceList{},
				Requested:               v1.ResourceList{},
			}
			Expect(k8sClient.Status().Update(context.TODO(), &actualQueue)).To(Succeed())
		}

		// Reconcile again to update project status
		result, err = reconciler.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		// Step 3: Verify project quota status is correctly cleared (not retaining old values)
		actualProject = kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())

		// NodePoolsQuotaStatuses should still exist but with empty resource lists
		Expect(len(actualProject.Status.NodePoolsQuotaStatuses)).To(Equal(len(actualProject.Spec.Queues)))
		for _, quotaStatus := range actualProject.Status.NodePoolsQuotaStatuses {
			Expect(quotaStatus.QueueStatus.Allocated).To(BeEmpty(),
				"NodePool %s Allocated should be empty after clearing queue status", quotaStatus.NodePoolName)
			Expect(quotaStatus.QueueStatus.AllocatedNonPreemptible).To(BeEmpty(),
				"NodePool %s AllocatedNonPreemptible should be empty after clearing queue status", quotaStatus.NodePoolName)
			Expect(quotaStatus.QueueStatus.Requested).To(BeEmpty(),
				"NodePool %s Requested should be empty after clearing queue status", quotaStatus.NodePoolName)
		}

		// Aggregated QuotaStatus should also be empty
		Expect(actualProject.Status.QuotaStatus.Allocated).To(BeEmpty(),
			"Project QuotaStatus.Allocated should be empty after clearing all queue statuses")
		Expect(actualProject.Status.QuotaStatus.AllocatedNonPreemptible).To(BeEmpty(),
			"Project QuotaStatus.AllocatedNonPreemptible should be empty after clearing all queue statuses")
		Expect(actualProject.Status.QuotaStatus.Requested).To(BeEmpty(),
			"Project QuotaStatus.Requested should be empty after clearing all queue statuses")
	})

	It("Reconciles an existing Project", func() {

		// Given
		sanity := kaiv1alpha1.Project{}
		Expect(k8sClient.Create(context.TODO(), &project)).To(Succeed())
		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).To(Succeed())
		Expect(sanity.Name).To(Equal(project.Name))

		// HandleResource project and sanity check
		result, err := reconciler.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())
		sanity = kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).To(Succeed())
		Expect(sanity.Spec.Parent).To(Equal(project.Spec.Parent))

		// When
		modifiedProject.ResourceVersion = sanity.ResourceVersion
		Expect(k8sClient.Update(context.TODO(), &modifiedProject)).To(Succeed())
		sanity = kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).To(Succeed())
		Expect(sanity.Name).To(Equal(project.Name))
		Expect(sanity.Spec.Parent).To(Equal(modifiedProject.Spec.Parent))

		result, err = reconciler.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		// Then

		By("Calling all handlers", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Name).To(Equal(project.Name))

			for _, queue := range actualProject.Spec.Queues {
				actualQueue := kaiv2.Queue{}

				Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: queue.Name}, &actualQueue)).To(Succeed())
				Expect(actualQueue.Name).To(Equal(queue.Name))
				Expect(actualQueue.OwnerReferences[0].UID).To(Equal(actualProject.UID))
				if queue.Nodepool != config.Get().DefaultNodepoolName {
					Expect(actualQueue.Labels[NodePoolLabelKey]).To(Equal(queue.Nodepool))
				} else {
					Expect(actualQueue.Labels).ToNot(HaveKey(NodePoolLabelKey))
				}
			}

			actualNamespace := v1.Namespace{}
			Expect(reconciler.Get(context.TODO(), client.ObjectKey{Name: newNamespaceName}, &actualNamespace)).To(Succeed())
			Expect(actualNamespace.Name).To(Equal(newNamespaceName))
			Expect(actualProject.Status.Namespace).To(Equal(newNamespaceName))
		})

	})

	It("Reconciles a Project which is being deleted", func() {
		// Given
		deletedProject := project.DeepCopy()
		deletedProject.DeletionTimestamp = &now

		// Without the other finalizer the fake client will remove the project when we remove our finalizer
		// https://github.com/kubernetes-sigs/controller-runtime/pull/1399
		deletedProject.Finalizers = []string{config.FinalizerName(), "other-finalizer"}
		// Fake client can no longer create objects with deleted timestamp - they have to be provided at Build()
		k8sClientWithDeletedProject := fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiv1alpha1.Project{}).
			WithStatusSubresource(&kaiv2.Queue{}).
			WithObjects(deletedProject).
			Build()

		reconcilerWithDeletedObjects := NewProjectReconciler(k8sClientWithDeletedProject, k8sClientWithDeletedProject, scheme, projectEvents, projectConfig)
		Expect(reconcilerWithDeletedObjects.Create(context.TODO(), &namespace)).To(Succeed())
		Expect(reconcilerWithDeletedObjects.Create(context.TODO(), &queue)).To(Succeed())

		// Sanity
		sanityProject := kaiv1alpha1.Project{}
		sanityQueue := kaiv2.Queue{}
		sanityNamespace := v1.Namespace{}
		Expect(reconcilerWithDeletedObjects.GetProject(projectReconcileRequest, &sanityProject)).To(Succeed())
		Expect(sanityProject.Name).To(Equal(project.Name))
		Expect(reconcilerWithDeletedObjects.Get(context.TODO(), client.ObjectKey{Name: queue.Name}, &sanityQueue)).To(Succeed())
		Expect(reconcilerWithDeletedObjects.Get(context.TODO(), client.ObjectKey{Name: TestNamespace.Name}, &sanityNamespace)).To(Succeed())

		// When
		result, err := reconcilerWithDeletedObjects.Reconcile(context.TODO(), projectReconcileRequest)
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		By("Calling the Finalizer", func() {
			Expect(reconcilerWithDeletedObjects.Get(context.TODO(), client.ObjectKey{Name: project.Name}, &sanityQueue)).ToNot(Succeed())
			// Project is not deleted by us, k8s api initiates the deletion and we just finalize
			Expect(reconcilerWithDeletedObjects.GetProject(projectReconcileRequest, &sanityProject)).To(Succeed())
		})

		By("Removing the reconciler from the Finalizers list", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconcilerWithDeletedObjects.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Finalizers).To(HaveLen(1))
		})

		By("Project Status is correct after finalization", func() {
			actualProject := kaiv1alpha1.Project{}
			Expect(reconcilerWithDeletedObjects.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Status.Phase).To(Equal(kaiv1alpha1.Ready))
			Expect(actualProject.Status.Message).To(Equal(""))
			Expect(len(actualProject.Status.Conditions)).ToNot(Equal(0))
		})
	})

	// Meaning the reconciler has logic to differentiate between a 'true' delete event (which is actually an update with deletion time stamp)
	// and the event that is received after finalization is complete
	It("Does nothing when Project is not found", func() {

		// Given
		sanity := kaiv1alpha1.Project{}
		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).ToNot(Succeed())
		Expect(sanity).To(BeZero())

		//When
		result, err := reconciler.Reconcile(context.TODO(), projectReconcileRequest)

		// Then
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		Expect(reconciler.GetProject(projectReconcileRequest, &sanity)).ToNot(Succeed())
		Expect(sanity).To(BeZero())
	})

	When("Reconcile fails", func() {
		var (
			badK8sClient  client.Client
			badReconciler *ProjectReconciler
			badScheme     *runtime.Scheme
		)
		BeforeEach(func() {
			badScheme = runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(badScheme)).Should(Succeed())
			Expect(kaiv1alpha1.AddToScheme(badScheme)).Should(Succeed())
			// not adding schedulingv2alpha2 to scheme - so queue handler should fail
			badK8sClient = fake.NewClientBuilder().
				WithScheme(badScheme).
				WithStatusSubresource(&kaiv1alpha1.Project{}).
				Build()

			config := config.NewProjectReconcilerConfig()
			badReconciler = NewProjectReconciler(badK8sClient, badK8sClient, badScheme, projectEvents, config)
		})

		It("Queue handler fails - Project NotReady", func() {
			// Given
			Expect(badK8sClient.Create(context.TODO(), &project)).To(Succeed())
			Expect(badReconciler.Create(context.TODO(), &namespace)).To(Succeed())

			// When
			_, err := badReconciler.Reconcile(context.TODO(), projectReconcileRequest)
			Expect(err).To(HaveOccurred())

			actualProject := kaiv1alpha1.Project{}
			Expect(badReconciler.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Status.Phase).To(Equal(kaiv1alpha1.NotReady))
			Expect(actualProject.Status.Message).To(Equal(string(handlers.QueuesHandlerFailed)))

			Expect(len(actualProject.Status.Conditions)).ToNot(Equal(0))
			var queuesCondition *kaiv1alpha1.ProjectCondition = nil
			for i := range actualProject.Status.Conditions {
				if actualProject.Status.Conditions[i].Type == kaiv1alpha1.QueuesReady {
					queuesCondition = &actualProject.Status.Conditions[i]
					break
				}
			}
			Expect(queuesCondition).ToNot(BeNil())
			Expect(queuesCondition.Status).To(Equal(v1.ConditionFalse))
			Expect(queuesCondition.Reason).To(Equal(string(handlers.QueuesHandlerFailed)))
			Expect(queuesCondition.Message).ToNot(Equal(""))
		})

		It("Fails to Reconcile a Project which is being deleted", func() {
			// Given
			deletedProject := project.DeepCopy()
			deletedProject.DeletionTimestamp = &now

			// Without the other finalizer the fake client will remove the project when we remove our finalizer
			// https://github.com/kubernetes-sigs/controller-runtime/pull/1399
			deletedProject.Finalizers = []string{config.FinalizerName(), "other-finalizer"}
			// Fake client can no longer create objects with deleted timestamp - they have to be provided at Build()
			k8sClientWithDeletedProject := fake.NewClientBuilder().
				WithScheme(badScheme).
				WithStatusSubresource(&kaiv1alpha1.Project{}).
				WithObjects(deletedProject).
				Build()

			reconcilerWithDeletedObjects := NewProjectReconciler(k8sClientWithDeletedProject, k8sClientWithDeletedProject, scheme, projectEvents, projectConfig)
			Expect(badReconciler.Create(context.TODO(), &namespace)).To(Succeed())

			_, err := reconcilerWithDeletedObjects.Reconcile(context.TODO(), projectReconcileRequest)
			Expect(err).To(HaveOccurred())

			actualProject := kaiv1alpha1.Project{}
			Expect(reconcilerWithDeletedObjects.GetProject(projectReconcileRequest, &actualProject)).To(Succeed())
			Expect(actualProject.Status.Phase).To(Equal(kaiv1alpha1.ProjectPhase("Deleting")))
			Expect(actualProject.Status.Message).To(Equal(string(deletion.QueuesDeletionHandlerFailed)))

			Expect(len(actualProject.Status.Conditions)).ToNot(Equal(0))
			var queuesCondition *kaiv1alpha1.ProjectCondition = nil
			for i := range actualProject.Status.Conditions {
				if actualProject.Status.Conditions[i].Type == kaiv1alpha1.QueuesReady {
					queuesCondition = &actualProject.Status.Conditions[i]
					break
				}
			}
			Expect(queuesCondition).ToNot(BeNil())
			Expect(queuesCondition.Status).To(Equal(v1.ConditionFalse))
			Expect(queuesCondition.Reason).To(Equal(string(deletion.QueuesDeletionHandlerFailed)))
			Expect(queuesCondition.Message).ToNot(Equal(""))
		})
	})

	When("Map events to projects correctly", func() {
		var testRoleBinding *rbacv1.RoleBinding
		var testNamespace *v1.Namespace

		BeforeEach(func() {
			testNamespace = TestNamespace.DeepCopy()
			err := k8sClient.Create(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())

			testRoleBinding = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-role-binding",
					Namespace: TestNamespace.Name,
				},
			}
		})

		It("Maps RoleBinding correctly", func() {
			// validate mapping from namespace label
			requests := reconciler.MapRoleBindingToProjectEvent(context.Background(), testRoleBinding)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.Name).To(Equal(TestProject.Name))

			// validate mapping from owner reference
			testRoleBinding.OwnerReferences = []metav1.OwnerReference{
				{
					Kind: "Project",
					Name: TestProject.Name,
				},
			}
			testRoleBinding.Namespace = "some-other-namespaceee"
			requests = reconciler.MapRoleBindingToProjectEvent(context.Background(), testRoleBinding)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.Name).To(Equal(TestProject.Name))

			// validate failed to find matching project
			testRoleBinding.OwnerReferences = []metav1.OwnerReference{}
			requests = reconciler.MapRoleBindingToProjectEvent(context.Background(), testRoleBinding)
			Expect(requests).To(HaveLen(0))
		})

		It("Maps Namespace correctly", func() {
			// validate mapping from namespace label
			requests := reconciler.MapNamespaceToProjectEvent(context.Background(), testNamespace)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.Name).To(Equal(TestProject.Name))

			// validate failed to find matching project
			testNamespace.Name = "some-other-namespaceee"
			testNamespace.Labels[config.Get().NamespaceProjectLabelKey] = ""
			requests = reconciler.MapNamespaceToProjectEvent(context.Background(), testNamespace)
			Expect(requests).To(HaveLen(0))
		})

		It("Maps the configured role-bindings ConfigMap to every Project", func() {
			projectConfig.RoleBindingsCm = "runai-project-controller-rolebindings-plugin"
			projectConfig.RoleBindingsCmNamespace = "runai"

			firstProject := TestProject.DeepCopy()
			secondProject := TestProject.DeepCopy()
			secondProject.Name = "second-project"
			overriddenProject := TestProject.DeepCopy()
			overriddenProject.Name = "manually-overridden-project"
			overriddenProject.Labels = map[string]string{RunaiResourceManualOverrideLabel: "true"}
			Expect(k8sClient.Create(context.Background(), firstProject)).To(Succeed())
			Expect(k8sClient.Create(context.Background(), secondProject)).To(Succeed())
			Expect(k8sClient.Create(context.Background(), overriddenProject)).To(Succeed())

			requests := reconciler.MapRoleBindingsConfigMapToProjectEvents(context.Background(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      projectConfig.RoleBindingsCm,
					Namespace: projectConfig.RoleBindingsCmNamespace,
				},
			})

			Expect(requests).To(ConsistOf(
				ctrl.Request{NamespacedName: client.ObjectKey{Name: firstProject.Name}},
				ctrl.Request{NamespacedName: client.ObjectKey{Name: secondProject.Name}},
			))
		})

		It("Does not map an unrelated ConfigMap to Projects", func() {
			projectConfig.RoleBindingsCm = "runai-project-controller-rolebindings-plugin"
			projectConfig.RoleBindingsCmNamespace = "runai"

			requests := reconciler.MapRoleBindingsConfigMapToProjectEvents(context.Background(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "runai"},
			})

			Expect(requests).To(BeEmpty())
		})
	})
})

func validateQueueStatusEqual(projQueueStatus *kaiv2.QueueStatus, queueStatus *kaiv2.QueueStatus) {
	Expect(projQueueStatus.Conditions).To(ConsistOf(queueStatus.Conditions))
	Expect(projQueueStatus.ChildQueues).To(Equal(queueStatus.ChildQueues))

	validateResourcesEqual(projQueueStatus.Allocated, queueStatus.Allocated,
		"Allocated")
	validateResourcesEqual(projQueueStatus.AllocatedNonPreemptible, queueStatus.AllocatedNonPreemptible,
		"AllocatedNonPreemptible")
	validateResourcesEqual(projQueueStatus.Requested, queueStatus.Requested,
		"Requested")
}

func validateResourcesEqual(actual v1.ResourceList, expected v1.ResourceList, fieldName string) {
	Expect(len(actual)).To(Equal(len(expected)))
	Expect(len(actual)).To(BeNumerically(">", 0))

	for key, expectedVal := range expected {
		Expect(actual).To(HaveKey(key))
		actualVal, found := actual[key]
		Expect(found).To(BeTrue(), "project status field <%s> to have key <%s>", fieldName, key)
		Expect(actualVal.Cmp(expectedVal)).To(Equal(0),
			"project status field <%s/%s> is:\n%+v\nshould have been:\n%+v",
			fieldName, key, actualVal, expectedVal)
	}
}
