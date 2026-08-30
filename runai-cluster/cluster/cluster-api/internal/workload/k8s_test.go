package workload

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	distinfwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Workload", func() {
	var (
		ctrl              *gomock.Controller
		mockClient        client.Client
		k8sLocator        *K8sLocator
		ctx               context.Context
		workloadName      string
		namespace         string
		expectedNamespace string
	)
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))   // core
	utilruntime.Must(iwv2.AddToScheme(scheme))     // interactive workload
	utilruntime.Must(twv2.AddToScheme(scheme))     // training workload\
	utilruntime.Must(dwv2.AddToScheme(scheme))     // distributed workload
	utilruntime.Must(v1alpha1.AddToScheme(scheme)) // external workload
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		k8sLocator = NewK8sLocator(mockClient)
		ctx = context.Background()
		workloadName = "test-workload"
		namespace = "test-project"
		expectedNamespace = "translated-namespace"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetK8sResource", func() {
		It("should return InteractiveWorkload for Workspace workload type", func() {
			resource, err := GetK8sResource(api.Workspace)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&iwv2.InteractiveWorkload{}))
		})

		It("should return TrainingWorkload for Training workload type", func() {
			resource, err := GetK8sResource(api.Training)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&twv2.TrainingWorkload{}))
		})

		It("should return DistributedWorkload for Distributed workload type", func() {
			resource, err := GetK8sResource(api.Distributed)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&dwv2.DistributedWorkload{}))
		})

		It("should return InferenceWorkload for Inference workload type", func() {
			resource, err := GetK8sResource(api.Inference)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&infwv2.InferenceWorkload{}))
		})

		It("should return DistributedInferenceWorkload for Distributed Inference workload type", func() {
			resource, err := GetK8sResource(api.DistributedInference)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&distinfwv2.DistributedInferenceWorkload{}))
		})

		It("should return ExternalWorkload for External workload type", func() {
			resource, err := GetK8sResource(api.External)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(BeAssignableToTypeOf(&v1alpha1.ExternalWorkload{}))
		})

		It("should return an error for unknown workload type", func() {
			resource, err := GetK8sResource("unknown")
			Expect(err).To(HaveOccurred())
			Expect(resource).To(BeNil())
		})
	})

	Describe("Locate", func() {
		It("should locate the workload successfully", func() {
			expectedResource := &iwv2.InteractiveWorkload{}
			expectedResource.SetName(workloadName)
			expectedResource.SetNamespace(expectedNamespace)

			Expect(mockClient.Create(ctx, expectedResource)).NotTo(HaveOccurred())
			resource, err := k8sLocator.Locate(ctx, api.Workspace, workloadName, expectedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(Equal(expectedResource))
		})

		It("should locate the external workload successfully", func() {
			expectedResource := &v1alpha1.ExternalWorkload{}
			expectedResource.SetName(workloadName)
			expectedResource.SetNamespace(expectedNamespace)

			Expect(mockClient.Create(ctx, expectedResource)).NotTo(HaveOccurred())
			resource, err := k8sLocator.Locate(ctx, api.External, workloadName, expectedNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource).To(Equal(expectedResource))
		})

		It("should return an error if getting the workload fails", func() {
			_, err := k8sLocator.Locate(ctx, api.Workspace, workloadName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})
})
