package distributedinference

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	resourcemanager "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	workloadConstants "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

// Simple mock implementations for testing
type MockResourceManager struct {
	// Return values for default UpsertResources
	UpsertResourcesIssues     []*issue.Issue
	UpsertResourcesIsConflict bool
	UpsertResourcesError      error

	// Track calls
	UpsertResourcesCalled    bool
	UpsertResourcesCallCount int
	LastWorkloadObjectMeta   metav1.ObjectMeta
	LastWorkloadSubResources []client.Object

	// Optional override for UpsertResources
	UpsertResourcesFunc func(ctx context.Context, workloadObjectMeta metav1.ObjectMeta, workloadSubResources []client.Object,
		ownedResources *[]common.ObservedResource, workloadConditions *[]common.WorkloadCondition, options resourcemanager.ResourceOptions) ([]*issue.Issue, bool)
}

func (m *MockResourceManager) UpsertResources(ctx context.Context, workloadObjectMeta metav1.ObjectMeta, workloadSubResources []client.Object,
	ownedResources *[]common.ObservedResource, workloadConditions *[]common.WorkloadCondition, options resourcemanager.ResourceOptions) ([]*issue.Issue, bool) {
	if m.UpsertResourcesFunc != nil {
		return m.UpsertResourcesFunc(ctx, workloadObjectMeta, workloadSubResources, ownedResources, workloadConditions, options)
	}
	m.UpsertResourcesCalled = true
	m.UpsertResourcesCallCount++
	m.LastWorkloadObjectMeta = workloadObjectMeta
	m.LastWorkloadSubResources = workloadSubResources
	return m.UpsertResourcesIssues, m.UpsertResourcesIsConflict
}

type MockDistributedInferenceBuilder struct {
	// Return values for BuildLws
	BuildLwsResult *lwsv1.LeaderWorkerSet
	BuildLwsError  error

	// Return values for BuildSubResources
	BuildSubResourcesResult []client.Object
	BuildSubResourcesError  error

	// Return values for GetPreview
	GetPreviewResult []client.Object
	GetPreviewError  error

	// Track calls
	BuildLwsCalled          bool
	BuildSubResourcesCalled bool
	GetPreviewCalled        bool
}

func (m *MockDistributedInferenceBuilder) BuildLws(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) (*lwsv1.LeaderWorkerSet, error) {
	m.BuildLwsCalled = true
	return m.BuildLwsResult, m.BuildLwsError
}

func (m *MockDistributedInferenceBuilder) BuildSubResources(ctx context.Context, workload *dinfw.DistributedInferenceWorkload, owner client.Object) ([]client.Object, error) {
	m.BuildSubResourcesCalled = true
	return m.BuildSubResourcesResult, m.BuildSubResourcesError
}

func (m *MockDistributedInferenceBuilder) GetPreview(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	m.GetPreviewCalled = true
	return m.GetPreviewResult, m.GetPreviewError
}

type MockDistributedStatusUpdater struct {
	// Return values for Update
	UpdateError error

	// Track calls
	UpdateCalled         bool
	LastWorkload         *dinfw.DistributedInferenceWorkload
	LastOriginalWorkload *dinfw.DistributedInferenceWorkload
	LastBuiltResources   []client.Object
	LastBuildErr         *error
}

func (m *MockDistributedStatusUpdater) Update(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	originalWorkload *dinfw.DistributedInferenceWorkload, builtResources []client.Object, buildErr *error) error {
	m.UpdateCalled = true
	m.LastWorkload = workload
	m.LastOriginalWorkload = originalWorkload
	m.LastBuiltResources = builtResources
	if buildErr != nil {
		// make a copy of buildErr to LastBuildErr
		buildErrCopy := *buildErr
		m.LastBuildErr = &buildErrCopy
	}
	return m.UpdateError
}

var _ = Describe("Controller", func() {
	var (
		ctx                 context.Context
		controller          *Controller
		mockResourceManager *MockResourceManager
		mockBuilder         *MockDistributedInferenceBuilder
		mockStatusUpdater   *MockDistributedStatusUpdater
		k8sClient           client.Client
		scheme              *runtime.Scheme
		workload            *dinfw.DistributedInferenceWorkload
		namespace           *corev1.Namespace
		req                 ctrl.Request
	)

	BeforeEach(func() {
		// Setup context
		ctx = context.Background()
		ctx = workloadproject.SetProjectInContext(ctx, "test-project")

		// Setup mocks
		mockResourceManager = &MockResourceManager{}
		mockBuilder = &MockDistributedInferenceBuilder{}
		mockStatusUpdater = &MockDistributedStatusUpdater{}

		// Setup scheme and client
		scheme = runtime.NewScheme()
		_ = dinfw.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		_ = netv1.AddToScheme(scheme)
		_ = lwsv1.AddToScheme(scheme)
		_ = gatewayv1.Install(scheme)

		// Create test workload
		workload = &dinfw.DistributedInferenceWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workload",
				Namespace: "runai-test-project",
				UID:       types.UID("test-uid"),
			},
			Spec: dinfp.DistributedInferenceSpec{},
		}

		namespace = resource.CreateNamespace("runai-test-project", "test-project")

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(workload, namespace).Build()

		// Set up environment variable
		_ = os.Setenv(workloadConstants.EnvVarClusterVersion, "1.0.0")

		// Create controller with mocks
		controller = &Controller{
			workloadType:    "DistributedInferenceWorkload",
			k8sClient:       k8sClient,
			isOpenshift:     false,
			resourceManager: mockResourceManager,
			builder:         mockBuilder,
			statusUpdater:   mockStatusUpdater,
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      workload.Name,
				Namespace: workload.Namespace,
			},
		}
	})

	AfterEach(func() {
		_ = os.Unsetenv(workloadConstants.EnvVarClusterVersion)
	})

	Describe("Reconcile", func() {
		Context("when workload exists and everything succeeds", func() {
			It("should reconcile successfully", func() {
				// Setup mock responses
				mockLws := &lwsv1.LeaderWorkerSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-lws",
						Namespace: workload.Namespace,
					},
				}
				mockBuilder.BuildLwsResult = mockLws
				mockBuilder.BuildLwsError = nil

				mockSubResources := []client.Object{
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-service",
							Namespace: workload.Namespace,
						},
					},
				}
				mockBuilder.BuildSubResourcesResult = mockSubResources
				mockBuilder.BuildSubResourcesError = nil

				// Resource manager should succeed
				mockResourceManager.UpsertResourcesIssues = []*issue.Issue{}
				mockResourceManager.UpsertResourcesIsConflict = false

				// Status updater should succeed
				mockStatusUpdater.UpdateError = nil

				// Execute
				result, err := controller.Reconcile(ctx, req)

				// Verify
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify calls were made
				Expect(mockBuilder.BuildLwsCalled).To(BeTrue())
				Expect(mockBuilder.BuildSubResourcesCalled).To(BeTrue())
				Expect(mockResourceManager.UpsertResourcesCalled).To(BeTrue())
				Expect(mockResourceManager.UpsertResourcesCallCount).To(Equal(2)) // LWS + SubResources
				Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
				Expect(mockStatusUpdater.LastBuildErr).To(BeNil())
			})
		})

		Context("when workload does not exist", func() {
			It("should return without error", func() {
				// Create client without the workload
				k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
				controller.k8sClient = k8sClient

				req.Name = "non-existent"

				result, err := controller.Reconcile(ctx, req)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify no calls were made to mocks
				Expect(mockBuilder.BuildLwsCalled).To(BeFalse())
				Expect(mockResourceManager.UpsertResourcesCalled).To(BeFalse())
				Expect(mockStatusUpdater.UpdateCalled).To(BeFalse())
			})
		})

		Context("when workload is marked for deletion", func() {
			It("should return without error", func() {
				// Mark workload for deletion
				now := metav1.Now()
				workload.DeletionTimestamp = &now
				workload.Finalizers = []string{"test-finalizer"} // required for fake client
				namespace = resource.CreateNamespace("runai-test-project", "test-project")
				k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(workload, namespace).Build()
				controller.k8sClient = k8sClient

				result, err := controller.Reconcile(ctx, req)

				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify no calls were made to mocks
				Expect(mockBuilder.BuildLwsCalled).To(BeFalse())
				Expect(mockResourceManager.UpsertResourcesCalled).To(BeFalse())
				Expect(mockStatusUpdater.UpdateCalled).To(BeFalse())
			})
		})

		Context("when LWS build fails", func() {
			It("should handle build error properly", func() {
				buildError := fmt.Errorf("failed to build LWS")
				mockBuilder.BuildLwsResult = nil
				mockBuilder.BuildLwsError = buildError

				// Mock preview resources for failed reconcile
				mockPreviewResources := []client.Object{}
				mockBuilder.GetPreviewResult = mockPreviewResources
				mockBuilder.GetPreviewError = nil

				// Status updater should be called with build error
				mockStatusUpdater.UpdateError = nil

				result, err := controller.Reconcile(ctx, req)

				// Should return without error for build errors (no requeue)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify calls
				Expect(mockBuilder.BuildLwsCalled).To(BeTrue())
				Expect(mockBuilder.GetPreviewCalled).To(BeTrue())
				Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
				Expect(mockStatusUpdater.LastBuildErr).ToNot(BeNil())
				Expect(*mockStatusUpdater.LastBuildErr).To(Equal(buildError))
			})
		})

		Context("when cluster version is not set", func() {
			It("should fail during label and annotation setting", func() {
				err := os.Unsetenv(workloadConstants.EnvVarClusterVersion)
				Expect(err).ToNot(HaveOccurred())

				// Status updater should be called with error
				mockStatusUpdater.UpdateError = nil

				result, err := controller.Reconcile(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify status updater was called with error
				Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
				Expect(mockStatusUpdater.LastBuildErr).To(BeNil())
			})
		})
	})

	Describe("buildAndUpsertResources", func() {
		It("should build and upsert LWS first, then sub resources", func() {
			// Setup mock responses
			mockLws := &lwsv1.LeaderWorkerSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-lws",
					Namespace: workload.Namespace,
				},
			}
			mockBuilder.BuildLwsResult = mockLws
			mockBuilder.BuildLwsError = nil

			mockSubResources := []client.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: workload.Namespace,
					},
				},
			}
			mockBuilder.BuildSubResourcesResult = mockSubResources
			mockBuilder.BuildSubResourcesError = nil

			// Resource manager should succeed
			mockResourceManager.UpsertResourcesIssues = []*issue.Issue{}
			mockResourceManager.UpsertResourcesIsConflict = false

			resources, issues, isConflict := controller.buildAndUpsertResources(ctx, workload)

			Expect(issues).To(BeEmpty())
			Expect(isConflict).To(BeFalse())
			Expect(resources).To(HaveLen(2)) // LWS + Service

			// Verify calls
			Expect(mockBuilder.BuildLwsCalled).To(BeTrue())
			Expect(mockBuilder.BuildSubResourcesCalled).To(BeTrue())
			Expect(mockResourceManager.UpsertResourcesCallCount).To(Equal(2))
		})
	})

	Describe("reconcileSucceeded", func() {
		It("should call status updater with nil build error", func() {
			resources := []client.Object{}
			originalWorkload := workload.DeepCopy()

			mockStatusUpdater.UpdateError = nil

			result, err := controller.reconcileSucceeded(ctx, workload, originalWorkload, resources)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify status updater was called correctly
			Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
			Expect(mockStatusUpdater.LastWorkload).To(Equal(workload))
			Expect(mockStatusUpdater.LastOriginalWorkload).To(Equal(originalWorkload))
			Expect(mockStatusUpdater.LastBuiltResources).To(Equal(resources))
			Expect(mockStatusUpdater.LastBuildErr).To(BeNil())
		})
	})

	Describe("reconcileFailed", func() {
		It("should call status updater with build error", func() {
			buildError := fmt.Errorf("build failed")
			originalWorkload := workload.DeepCopy()
			resources := []client.Object{}

			mockStatusUpdater.UpdateError = nil

			result, err := controller.reconcileFailed(ctx, buildError, true, workload, originalWorkload, resources)

			// Build errors should not requeue
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify status updater was called with build error
			Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
			Expect(mockStatusUpdater.LastBuildErr).ToNot(BeNil())
			Expect(*mockStatusUpdater.LastBuildErr).To(Equal(buildError))
		})

		It("should get preview resources when none provided", func() {
			buildError := fmt.Errorf("build failed")
			originalWorkload := workload.DeepCopy()

			// Setup preview resources
			previewResources := []client.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "preview-service"},
				},
			}
			mockBuilder.GetPreviewResult = previewResources
			mockBuilder.GetPreviewError = nil

			mockStatusUpdater.UpdateError = nil

			result, err := controller.reconcileFailed(ctx, buildError, true, workload, originalWorkload, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify GetPreview was called and status updater received preview resources
			Expect(mockBuilder.GetPreviewCalled).To(BeTrue())
			Expect(mockStatusUpdater.UpdateCalled).To(BeTrue())
			Expect(mockStatusUpdater.LastBuiltResources).To(Equal(previewResources))
		})
	})

	Describe("upsertLws", func() {
		var (
			mockLws *lwsv1.LeaderWorkerSet
		)

		BeforeEach(func() {
			mockLws = &lwsv1.LeaderWorkerSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-lws",
					Namespace: workload.Namespace,
				},
			}
		})

		It("should retry on conflict error and return issues, true", func() {
			// Simulate conflict error on first call, then success
			callCount := 0
			mockResourceManager.UpsertResourcesFunc = func(ctx context.Context, workloadObjectMeta metav1.ObjectMeta, workloadSubResources []client.Object,
				ownedResources *[]common.ObservedResource, workloadConditions *[]common.WorkloadCondition, options resourcemanager.ResourceOptions) ([]*issue.Issue, bool) {
				callCount++
				if callCount < 3 {
					return []*issue.Issue{issue.New(mockLws, fmt.Errorf("conflict error"), true, "conflict", nil)}, true
				}
				return []*issue.Issue{}, false
			}
			issues, isConflict := controller.upsertLws(ctx, workload, mockLws)
			// Should succeed after retries (no issues, isConflict false)
			Expect(issues).To(BeNil())
			Expect(isConflict).To(BeFalse())
			mockResourceManager.UpsertResourcesFunc = nil
		})

		It("should return issues, false on non-conflict error", func() {
			mockResourceManager.UpsertResourcesIssues = []*issue.Issue{issue.New(mockLws, fmt.Errorf("some other error"), false, "other error", nil)}
			mockResourceManager.UpsertResourcesIsConflict = false

			issues, isConflict := controller.upsertLws(ctx, workload, mockLws)
			Expect(issues).ToNot(BeNil())
			Expect(isConflict).To(BeFalse())
		})
	})
})
